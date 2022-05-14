package rpc

import (
	"context"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/transport"
)

type requestStream struct {
	requestNumber int
	typ           ProcedureType

	sentCloseStream     bool
	sentCloseStreamLock sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	raw    MessageSender
}

func newRequestStream(ctx context.Context, number int, typ ProcedureType, raw MessageSender) *requestStream {
	ctx, cancel := context.WithCancel(ctx)

	rs := &requestStream{
		requestNumber: number,
		typ:           typ,

		ctx:    ctx,
		cancel: cancel,
		raw:    raw,
	}

	return rs
}

func (rs *requestStream) WriteMessage(body []byte) error {
	select {
	case <-rs.ctx.Done():
		return rs.ctx.Err()
	default:
	}

	// todo do this correctly? are the flags correct?
	flags, err := transport.NewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON)
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

func (rs *requestStream) CloseWithError(err error) error {
	rs.sentCloseStreamLock.Lock()
	defer rs.sentCloseStreamLock.Unlock()

	if rs.sentCloseStream {
		return errors.New("already sent close stream")
	}

	rs.cancel()
	rs.sentCloseStream = true
	return sendCloseStream(rs.raw, -rs.requestNumber, err)
}

func (rs *requestStream) Context() context.Context {
	return rs.ctx
}

func (rs *requestStream) RequestNumber() int {
	return rs.requestNumber
}

func (rs *requestStream) TerminatedByRemote() {
	rs.cancel()
}

func (rs *requestStream) HandleNewMessage(msg *transport.Message) error {
	if rs.typ != ProcedureTypeDuplex {
		return errors.New("illegal duplicate request number")
	}

	// todo pass msg to the handler

	return nil
}
