package rpc

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/transport"
)

type ResponseWriter interface {
	// WriteMessage sends a message over the underlying stream.
	WriteMessage(body []byte) error

	// CloseWithError terminates the underlying stream. Error is sent to the
	// other party. Error can be nil.
	CloseWithError(err error) error
}

type RequestHandler interface {
	// HandleRequest should respond to the provided request using the response
	// writer. Implementations must eventually call
	// ResponseWriter.CloseWithError. HandleRequest may block as it is executed
	// in a goroutine.
	HandleRequest(ctx context.Context, rw ResponseWriter, req *Request)
}

// RequestStreams is used for handling streams initiated by remote (for which
// incoming messages have positive request numbers).
type RequestStreams struct {
	streams       map[int]*RequestStream
	closedStreams map[int]struct{}
	streamsLock   sync.Mutex // guards streams and closedStreams

	handler RequestHandler

	raw    MessageSender
	logger logging.Logger
}

// NewRequestStreams creates new request streams which use the provided context
// to run the loop cleaning up closed streams. The lifecycle of this context
// should most likely be the same as the lifecycle of the underlying connection.
func NewRequestStreams(ctx context.Context, raw MessageSender, handler RequestHandler, logger logging.Logger) *RequestStreams {
	rs := &RequestStreams{
		raw:           raw,
		streams:       make(map[int]*RequestStream),
		closedStreams: make(map[int]struct{}),
		handler:       handler,
		logger:        logger.New("response_streams"),
	}
	go rs.cleanupLoop(ctx)
	return rs
}

// HandleIncomingRequest processes an incoming request. Returning an error from
// this function shuts down the entire connection.
func (s *RequestStreams) HandleIncomingRequest(ctx context.Context, msg *transport.Message) error {
	if !msg.Header.IsRequest() {
		return errors.New("passed a response")
	}

	s.streamsLock.Lock()
	defer s.streamsLock.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	requestNumber := msg.Header.RequestNumber()

	// Unfortunately we are not able to distinguish "start of stream" messages
	// from "part of stream" messages in the case of duplex connections. This is
	// a problem which exists at the design level of the Secure Scuttlebutt's
	// RPC protocol. Both the initial message in a duplex connection and
	// follow-up messages:
	// - have the stream flag set
	// - have the same request number
	// It is therefore impossible to distinguish them without really weird
	// payload-level heuristics or broken bookkeeping like the one applied here
	// which may eventually cause the program to run out of memory.
	//
	// Real scenario:
	// If we close a duplex connection and then a new message with the same
	// request number comes in we have to remember that we should discard it
	// instead of opening a new stream based on this message. We have no other
	// means to realise that this message is not a new request.
	//
	// This could be partially mitigated if we assumed that new request numbers
	// can only be larger than previous request numbers but protocol docs make
	// no such assumptions. Technically they don't also mention that request
	// numbers can't be reused but there is simply no way to implement the RPC
	// protocol without making that assumption.
	if _, ok := s.closedStreams[requestNumber]; ok {
		return nil
	}

	existingStream, ok := s.streams[requestNumber]
	if ok {
		if msg.Header.Flags().EndOrError() {
			existingStream.TerminatedByRemote()
			return nil
		}
		return existingStream.HandleNewMessage(msg)
	}

	return s.openNewRequestStream(ctx, msg)
}

// openNewRequestStream unmarshalls the request contained in the provided
// message and opens a new stream for that request. If the request is malformed
// then it is ignored. This is done because unfortunately many clients create
// requests which are malformed according to the protocol guide.
// See: http://dev.planetary.social/rpc
func (s *RequestStreams) openNewRequestStream(ctx context.Context, msg *transport.Message) error {
	requestNumber := msg.Header.RequestNumber()

	req, err := unmarshalRequest(msg)
	if err != nil {
		s.logger.WithError(err).Debug("ignoring malformed request")

		if sendCloseStreamErr := sendCloseStream(
			s.raw,
			-msg.Header.RequestNumber(),
			errors.New("received malformed request"),
		); sendCloseStreamErr != nil {
			if !errors.Is(sendCloseStreamErr, net.ErrClosed) {
				s.logger.WithError(sendCloseStreamErr).Error("error sending close stream message")
			}
		}

		return nil
	}

	rs, err := NewRequestStream(ctx, requestNumber, req.Type(), s.raw)
	if err != nil {
		return errors.Wrap(err, "could not create a request stream")
	}

	s.streams[requestNumber] = rs

	go s.handler.HandleRequest(rs.Context(), rs, req)

	return nil
}

func (s *RequestStreams) cleanupLoop(ctx context.Context) {
	for {
		s.cleanup()

		select {
		case <-time.After(10 * time.Millisecond):
			continue
		case <-ctx.Done():
			return
		}
	}
}

// cleanup removes request streams which were closed by either party and marks
// them as closed so that future incoming messages for this stream can be
// ignored. This cleanup doesn't need to happen right away as:
// - closing a stream many times is allowed
// - if someone sends repeated messages then HandleNewMessage method should be
//   able to handle it for async streams and the message will be rejected anyway
//   for all other streams
func (s *RequestStreams) cleanup() {
	s.streamsLock.Lock()
	defer s.streamsLock.Unlock()

	for _, rs := range s.streams {
		select {
		case <-rs.Context().Done():
			delete(s.streams, rs.RequestNumber())
			s.closedStreams[rs.RequestNumber()] = struct{}{}
		default:
			continue
		}
	}
}

func unmarshalRequest(msg *transport.Message) (*Request, error) {
	var requestBody RequestBody
	if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal the request body")
	}

	procedureName, err := NewProcedureName(requestBody.Name)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a procedure name")
	}

	procedureType := decodeProcedureType(requestBody.Type)

	req, err := NewRequest(
		procedureName,
		procedureType,
		requestBody.Args,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a request")
	}

	return req, err
}

// sendCloseStream closes the specified stream. Number is put directly into the
// request number field of the sent message. If you are closing a stream that
// you initiated then number will be a positive value, if you are closing a
// stream that a peer initiated then number will be a negative value.
func sendCloseStream(raw MessageSender, number int, errToSent error) error {
	// todo do this correctly? are the flags correct?
	flags, err := transport.NewMessageHeaderFlags(true, true, transport.MessageBodyTypeJSON)
	if err != nil {
		return errors.Wrap(err, "could not create message header flags")
	}

	var content []byte
	if errToSent == nil {
		content = []byte("true") // todo why true, is there any reason for this? do we have to send something specific? is this documented?
	} else {
		var mErr error
		content, mErr = json.Marshal(struct {
			Error string `json:"error"`
		}{errToSent.Error()})
		if mErr != nil {
			panic(mErr) // tests would have caught this eg. TestPrematureTerminationByRemote
		}
	}

	header, err := transport.NewMessageHeader(flags, uint32(len(content)), int32(number))
	if err != nil {
		return errors.Wrap(err, "could not create a message header")
	}

	msg, err := transport.NewMessage(header, content)
	if err != nil {
		return errors.Wrap(err, "could not create a message")
	}

	if err := raw.Send(&msg); err != nil {
		return errors.Wrap(err, "could not send a message")
	}

	return nil
}
