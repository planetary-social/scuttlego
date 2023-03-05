package rpc

import (
	"context"
	"net"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
)

type onLocalCloseFn func(rs *RequestStream)

type RequestStream struct {
	requestNumber int
	typ           ProcedureType

	sentCloseStream     bool
	sentCloseStreamLock sync.Mutex

	raw MessageSender

	ctx           context.Context
	cancelContext context.CancelFunc

	onLocalClose onLocalCloseFn

	incomingMessages chan IncomingMessage
}

func NewRequestStream(ctx context.Context, onLocalClose onLocalCloseFn, number int, typ ProcedureType, raw MessageSender) (*RequestStream, error) {
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

		raw: raw,

		ctx:           ctx,
		cancelContext: cancel,
		onLocalClose:  onLocalClose,

		incomingMessages: make(chan IncomingMessage),
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

	rs.onLocalClose(rs)

	rs.sentCloseStream = true

	if err := sendCloseStream(rs.raw, -rs.requestNumber, err); err != nil {
		if errors.Is(err, net.ErrClosed) {
			return nil
		}
		return errors.Wrap(err, "failed to send close stream")
	}
	return nil
}

func (rs *RequestStream) IncomingMessages() (<-chan IncomingMessage, error) {
	if rs.typ != ProcedureTypeDuplex {
		return nil, errors.New("only duplex streams can receive messages")
	}
	return rs.incomingMessages, nil
}

func (rs *RequestStream) RequestNumber() int {
	return rs.requestNumber
}

func (rs *RequestStream) HandleNewMessage(msg transport.Message) error {
	if rs.typ != ProcedureTypeDuplex {
		return errors.New("only duplex streams can receive messages")
	}

	select {
	case <-rs.ctx.Done():
	case rs.incomingMessages <- IncomingMessage{Body: msg.Body}:
	}
	return nil
}
