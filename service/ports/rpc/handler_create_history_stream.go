package rpc

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
)

// CreateHistoryStreamQueryHandler is here to make testing easier. See docs for
// the CreateHistoryStream application query.
type CreateHistoryStreamQueryHandler interface {
	Handle(ctx context.Context, query queries.CreateHistoryStream)
}

type HandlerCreateHistoryStream struct {
	q      CreateHistoryStreamQueryHandler
	logger logging.Logger
}

func NewHandlerCreateHistoryStream(
	q CreateHistoryStreamQueryHandler,
	logger logging.Logger,
) *HandlerCreateHistoryStream {
	return &HandlerCreateHistoryStream{
		q:      q,
		logger: logger.New("create_history_stream"),
	}
}

func (h *HandlerCreateHistoryStream) Procedure() rpc.Procedure {
	return messages.CreateHistoryStreamProcedure
}

func (h *HandlerCreateHistoryStream) Handle(ctx context.Context, s mux.CloserStream, req *rpc.Request) {
	args, err := messages.NewCreateHistoryStreamArgumentsFromBytes(req.Arguments())
	if err != nil {
		if closeErr := s.CloseWithError(err); closeErr != nil {
			h.logger.WithError(closeErr).Debug("could not close the stream")
		}
		return
	}

	query := queries.CreateHistoryStream{
		Id:             args.Id(),
		Seq:            args.Sequence(),
		Limit:          args.Limit(),
		Live:           args.Live(),
		Old:            args.Old(),
		ResponseWriter: NewCreateHistoryStreamResponseWriter(args, s),
	}

	h.q.Handle(ctx, query)
}

type CreateHistoryStreamResponseWriter struct {
	args messages.CreateHistoryStreamArguments
	s    rpc.Stream
}

func NewCreateHistoryStreamResponseWriter(
	args messages.CreateHistoryStreamArguments,
	s rpc.Stream,
) *CreateHistoryStreamResponseWriter {
	return &CreateHistoryStreamResponseWriter{
		args: args,
		s:    s,
	}
}

func (rw CreateHistoryStreamResponseWriter) WriteMessage(msg message.Message) error {
	b, err := rw.createResponse(msg)
	if err != nil {
		return errors.Wrap(err, "could not create a response")
	}

	if err := rw.s.WriteMessage(b); err != nil {
		return errors.Wrap(err, "could not write the message")
	}

	return nil
}

func (rw CreateHistoryStreamResponseWriter) CloseWithError(err error) error {
	return rw.s.CloseWithError(err)
}

func (rw CreateHistoryStreamResponseWriter) createResponse(msg message.Message) ([]byte, error) {
	if rw.args.Keys() {
		// todo what is the timestamp used for? do we actually need to remember when we stored something?
		return messages.NewCreateHistoryStreamResponse(msg.Id(), msg.Raw(), time.Now()).MarshalJSON()
	}
	return msg.Raw().Bytes(), nil
}
