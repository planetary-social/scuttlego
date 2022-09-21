package ebt

import (
	"context"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type ResponseStreamAdapter struct {
	stream *rpc.ResponseStream
}

func NewResponseStreamAdapter(stream *rpc.ResponseStream) *ResponseStreamAdapter {
	return &ResponseStreamAdapter{stream: stream}
}

func (r *ResponseStreamAdapter) IncomingMessages(ctx context.Context) <-chan IncomingMessage {
	ch := make(chan IncomingMessage)
	go func() {
		defer close(ch)

		for resp := range r.stream.Channel() {
			if err := r.parseErr(resp.Err); err != nil {
				select {
				case <-ctx.Done():
					return
				case ch <- NewIncomingMessageWithErr(err):
					return
				}
			}

			incomingMessage, err := r.parse(resp.Value.Bytes())
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case ch <- NewIncomingMessageWithErr(err):
					return
				}
			}

			select {
			case <-ctx.Done():
				return
			case ch <- incomingMessage:
			}
		}
	}()
	return ch
}

func (r *ResponseStreamAdapter) parseErr(err error) error {
	if err != nil {
		if errors.Is(err, rpc.ErrEndOrErr) {
			return replication.ErrPeerDoesNotSupportEBT
		}
		return errors.Wrap(err, "stream returned an error")
	}
	return nil
}

func (r *ResponseStreamAdapter) parse(b []byte) (IncomingMessage, error) {
	var returnErr error

	note, err := messages.NewEbtReplicateNoteFromBytes(b)
	if err == nil {
		return NewIncomingMessageWithNote(note), nil
	}
	returnErr = multierror.Append(returnErr, errors.Wrap(err, "could not create a new note"))

	rawMessage, err := message.NewRawMessage(b)
	if err == nil {
		return NewIncomingMessageWithMesage(rawMessage), nil
	}
	returnErr = multierror.Append(returnErr, errors.Wrap(err, "could not create a new raw message"))

	return IncomingMessage{}, returnErr
}

func (r *ResponseStreamAdapter) SendNote(note messages.EbtReplicateNote) {
	panic("implement me")
}

func (r *ResponseStreamAdapter) SendMessage(msg *message.Message) {
	panic("implement me")
}
