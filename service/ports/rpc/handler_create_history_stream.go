package rpc

import (
	"context"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
)

const numWorkers = 10

type HandlerCreateHistoryStream struct {
	q CreateHistoryStreamQueryHandler

	requests     []IncomingRequest
	requestsLock sync.Mutex

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

func (h *HandlerCreateHistoryStream) Run(ctx context.Context) error {
	h.startWorkers(ctx)
	<-ctx.Done()
	return ctx.Err()
}

func (h *HandlerCreateHistoryStream) Handle(ctx context.Context, rw mux.ResponseWriterCloser, req *rpc.Request) {
	h.requestsLock.Lock()
	defer h.requestsLock.Unlock()

	h.requests = append(h.requests, IncomingRequest{
		Ctx: ctx,
		Rw:  rw,
		Req: req,
	})
}

func (h *HandlerCreateHistoryStream) sendMessage(args messages.CreateHistoryStreamArguments, msg message.Message, rw mux.ResponseWriter) error {
	b, err := h.createResponse(args, msg)
	if err != nil {
		return errors.Wrap(err, "could not create a response")
	}

	if err := rw.WriteMessage(b); err != nil {
		return errors.Wrap(err, "could not write the message")
	}

	return nil
}

func (h *HandlerCreateHistoryStream) createResponse(args messages.CreateHistoryStreamArguments, msg message.Message) ([]byte, error) {
	if args.Keys() {
		// todo what is the timestamp used for? do we actually need to remember when we stored something?
		return messages.NewCreateHistoryStreamResponse(msg.Id(), msg.Raw(), time.Now()).MarshalJSON()
	}
	return msg.Raw().Bytes(), nil
}

func (h *HandlerCreateHistoryStream) startWorkers(ctx context.Context) {
	for i := 0; i < numWorkers; i++ {
		go h.worker(ctx)
	}
}

func (h *HandlerCreateHistoryStream) worker(ctx context.Context) {
	for {
		req, ok := h.getNextRequest()
		if !ok {
			select {
			case <-time.After(10 * time.Millisecond):
				continue
			case <-ctx.Done():
				return
			}
		}

		if err := h.processRequest(req); err != nil {
			h.logger.WithError(err).Debug("error processing an incoming request")
		}
	}
}

func (h *HandlerCreateHistoryStream) getNextRequest() (IncomingRequest, bool) {
	h.requestsLock.Lock()
	defer h.requestsLock.Unlock()

	if len(h.requests) == 0 {
		return IncomingRequest{}, false
	}

	req := h.requests[0]
	h.requests = h.requests[:1]
	return req, true
}

func (h *HandlerCreateHistoryStream) processRequest(in IncomingRequest) error {
	args, err := messages.NewCreateHistoryStreamArgumentsFromBytes(in.Req.Arguments())
	if err != nil {
		if err := in.Rw.CloseWithError(errors.Wrap(err, "invalid arguments")); err != nil {
			h.logger.WithError(err).Debug("closing failed")
		}
		return errors.Wrap(err, "invalid arguments")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// register for live

	// send old

	// send remaining live

	// switch to live

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

type IncomingRequest struct {
	Ctx context.Context
	Rw  mux.ResponseWriterCloser
	Req *rpc.Request
}
