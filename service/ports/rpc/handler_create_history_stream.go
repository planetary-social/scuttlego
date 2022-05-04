package rpc

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/mux"
)

// CreateHistoryStreamQueryHandler is here to make testing easier. See docs for
// the CreateHistoryStream application query.
type CreateHistoryStreamQueryHandler interface {
	Handle(ctx context.Context, query queries.CreateHistoryStream) <-chan queries.MessageWithErr
}

type HandlerCreateHistoryStream struct {
	q CreateHistoryStreamQueryHandler
}

func NewHandlerCreateHistoryStream(q CreateHistoryStreamQueryHandler) *HandlerCreateHistoryStream {
	return &HandlerCreateHistoryStream{
		q: q,
	}
}

func (h HandlerCreateHistoryStream) Procedure() rpc.Procedure {
	return messages.CreateHistoryStreamProcedure
}

func (h HandlerCreateHistoryStream) Handle(ctx context.Context, rw mux.ResponseWriter, req *rpc.Request) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	args, err := messages.NewCreateHistoryStreamArgumentsFromBytes(req.Arguments())
	if err != nil {
		return errors.Wrap(err, "invalid arguments")
	}

	query := queries.CreateHistoryStream{
		Id:    args.Id(),
		Seq:   args.Sequence(),
		Limit: args.Limit(),
		Live:  args.Live(),
		Old:   args.Old(),
	}

	msgCh := h.q.Handle(ctx, query)

	for msgWithError := range msgCh {
		if msgWithError.Err != nil {
			return errors.Wrap(err, "query returned an error")
		}

		if err := h.sendMessage(args, msgWithError.Message, rw); err != nil {
			return errors.Wrap(err, "could not send a message")
		}
	}

	return nil
}

func (h HandlerCreateHistoryStream) sendMessage(args messages.CreateHistoryStreamArguments, msg message.Message, rw mux.ResponseWriter) error {
	b, err := h.createResponse(args, msg)
	if err != nil {
		return errors.Wrap(err, "could not create a response")
	}

	if err := rw.WriteMessage(b); err != nil {
		return errors.Wrap(err, "could not write the message")
	}

	return nil
}

func (h HandlerCreateHistoryStream) createResponse(args messages.CreateHistoryStreamArguments, msg message.Message) ([]byte, error) {
	if args.Keys() {
		// todo what is the timestamp used for? do we actually need to remember when we stored something?
		return messages.NewCreateHistoryStreamResponse(msg.Id(), msg.Raw(), time.Now()).MarshalJSON()
	}
	return msg.Raw().Bytes(), nil
}
