package rpc

import (
	"context"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
)

// CreateHistoryStreamQueryHandler is here to make testing easier. See docs for
// the CreateHistoryStream application query.
type CreateHistoryStreamQueryHandler interface {
	Handle(ctx context.Context, query queries.CreateHistoryStream) <-chan queries.MessageWithErr
}

type CreateHistoryStreamProcessor struct {
}

func (p CreateHistoryStreamProcessor) Handle(ctx context.Context, rw mux.ResponseWriterCloser, req *rpc.Request) {
	h.requestsLock.Lock()
	defer h.requestsLock.Unlock()

}

//h.requests = append(h.requests, IncomingCreateHistoryStreamRequest{
//	Ctx: ctx,
//	Rw:  rw,
//	Arguments:,
//})
//
//ctx, cancel := context.WithCancel(ctx)
//defer cancel()
//
//query := queries.CreateHistoryStream{
//	Id:    args.Id(),
//	Seq:   args.Sequence(),
//	Limit: args.Limit(),
//	Live:  args.Live(),
//	Old:   args.Old(),
//}
//
//msgCh := h.q.Handle(ctx, query)
//
//for msgWithError := range msgCh {
//	if msgWithError.Err != nil {
//		return errors.Wrap(err, "query returned an error")
//	}
//
//	if err := h.sendMessage(args, msgWithError.Message, rw); err != nil {
//		return errors.Wrap(err, "could not send a message")
//	}
//}
//
//return nil

//func (h HandlerCreateHistoryStream) sendMessage(args messages.CreateHistoryStreamArguments, msg message.Message, rw mux.ResponseWriter) error {
//	b, err := h.createResponse(args, msg)
//	if err != nil {
//		return errors.Wrap(err, "could not create a response")
//	}
//
//	if err := rw.WriteMessage(b); err != nil {
//		return errors.Wrap(err, "could not write the message")
//	}
//
//	return nil
//}
//
//func (h HandlerCreateHistoryStream) createResponse(args messages.CreateHistoryStreamArguments, msg message.Message) ([]byte, error) {
//	if args.Keys() {
//		// todo what is the timestamp used for? do we actually need to remember when we stored something?
//		return messages.NewCreateHistoryStreamResponse(msg.Id(), msg.Raw(), time.Now()).MarshalJSON()
//	}
//	return msg.Raw().Bytes(), nil
//}
