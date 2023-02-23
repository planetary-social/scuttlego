package rpc

import (
	"context"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
)

type RequestStream struct {
	requestNumber int
	typ           ProcedureType

	sentCloseStream     bool
	sentCloseStreamLock sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	raw    MessageSender

	incomingMessages       chan IncomingMessage
	incomingMessagesLock   sync.Mutex
	incomingMessagesClosed bool
}

func NewRequestStream(ctx context.Context, number int, typ ProcedureType, raw MessageSender) (*RequestStream, error) {
	if number <= 0 {
		return nil, errors.New("number must be positive")
	}

	if typ.IsZero() {
		return nil, errors.New("zero value of procedure type")
	}

	ctx, cancel := context.WithCancel(ctx)

	rs := &RequestStream{
		requestNumber: number,
		typ:           typ,

		ctx:    ctx,
		cancel: cancel,
		raw:    raw,

		incomingMessages: make(chan IncomingMessage),
	}

	if typ == ProcedureTypeDuplex {
		go func() {
			<-ctx.Done()

			rs.incomingMessagesLock.Lock()
			rs.incomingMessagesClosed = true
			close(rs.incomingMessages)
			rs.incomingMessagesLock.Unlock()
		}()
	}

	return rs, nil
}

func (rs *RequestStream) WriteMessage(body []byte, bodyType transport.MessageBodyType) error {
	if bodyType.IsZero() {
		return errors.New("zero value of body type")
	}

	select {
	case <-rs.ctx.Done():
		return rs.ctx.Err()
	default:
	}

	// todo is the stream flag correct?
	flags, err := transport.NewMessageHeaderFlags(true, false, bodyType)
	if err != nil {
		return errors.Wrap(err, "could not create message header flags")
	}

	header, err := transport.NewMessageHeader(flags, uint32(len(body)), int32(-rs.requestNumber))
	if err != nil {
		return errors.Wrap(err, "could not create a message header")
	}

	msg, err := transport.NewMessage(header, body)
	if err != nil {
		return errors.Wrap(err, "could not create a message")
	}

	if err := rs.raw.Send(&msg); err != nil {
		return errors.Wrap(err, "could not send a message")
	}

	return nil
}

func (rs *RequestStream) CloseWithError(err error) error {
	rs.sentCloseStreamLock.Lock()
	defer rs.sentCloseStreamLock.Unlock()

	if rs.sentCloseStream {
		return errors.New("already sent close stream")
	}

	rs.cancel()
	rs.sentCloseStream = true
	return sendCloseStream(rs.raw, -rs.requestNumber, err)
}

func (rs *RequestStream) IncomingMessages() (<-chan IncomingMessage, error) {
	if rs.typ != ProcedureTypeDuplex {
		return nil, errors.New("only duplex streams can receive messages")
	}
	return rs.incomingMessages, nil
}

func (rs *RequestStream) Context() context.Context {
	return rs.ctx
}

func (rs *RequestStream) RequestNumber() int {
	return rs.requestNumber
}

func (rs *RequestStream) TerminatedByRemote() {
	rs.cancel()
}

func (rs *RequestStream) HandleNewMessage(msg transport.Message) error {
	if rs.typ != ProcedureTypeDuplex {
		return errors.New("only duplex streams can receive messages")
	}

	rs.incomingMessagesLock.Lock()
	defer rs.incomingMessagesLock.Unlock()

	if rs.incomingMessagesClosed {
		return nil
	}

	select {
	case <-rs.ctx.Done():
	case rs.incomingMessages <- IncomingMessage{Body: msg.Body}:
	}
	return nil
}
