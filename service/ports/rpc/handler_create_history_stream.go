package rpc

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/app"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

type HandlerCreateHistoryStream struct {
	app app.Application
}

func NewHandlerCreateHistoryStream(app app.Application) *HandlerCreateHistoryStream {
	return &HandlerCreateHistoryStream{
		app: app,
	}
}

func (h HandlerCreateHistoryStream) Procedure() rpc.Procedure {
	return messages.CreateHistoryStreamProcedure
}

func (h HandlerCreateHistoryStream) Handle(req *rpc.Request, w *rpc.ResponseWriter) error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	query, err := h.newQuery(req)
	if err != nil {
		return errors.Wrap(err, "could not construct a query")
	}

	msgCh := h.app.Queries.CreateHistoryStream.Handle(ctx, query)

	for msgWithError := range msgCh {
		if msgWithError.Err != nil {
			return errors.Wrap(err, "query returned an error")
		}

		// todo write messages
	}

	return nil
}

func (h HandlerCreateHistoryStream) newQuery(req *rpc.Request) (queries.CreateHistoryStream, error) {
	args, err := messages.NewCreateHistoryStreamArgumentsFromBytes(req.Arguments())
	if err != nil {
		return queries.CreateHistoryStream{}, errors.Wrap(err, "invalid arguments")
	}

	live := false
	if args.Live() != nil {
		live = *args.Live()
	}

	old := true
	if args.Old() != nil {
		old = *args.Old()
	}

	return queries.CreateHistoryStream{
		Id:    args.Id(),
		Seq:   args.Sequence(),
		Limit: args.Limit(),
		Live:  live,
		Old:   old,
	}, nil
}
