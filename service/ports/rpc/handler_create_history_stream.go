package rpc

import (
	"github.com/boreq/errors"
	rpc2 "github.com/planetary-social/go-ssb/service/domain/network/rpc"
	"github.com/planetary-social/go-ssb/service/domain/network/rpc/messages"
	"time"
)

type HandlerCreateHistoryStream struct {
}

func NewHandlerCreateHistoryStream() *HandlerCreateHistoryStream {
	return &HandlerCreateHistoryStream{}
}

func (h HandlerCreateHistoryStream) Procedure() rpc2.Procedure {
	return messages.CreateHistoryStreamProcedure
}

func (h HandlerCreateHistoryStream) Handle(req *rpc2.Request, w *rpc2.ResponseWriter) error {
	_, err := messages.NewCreateHistoryStreamArgumentsFromBytes(req.Arguments())
	if err != nil {
		return errors.Wrap(err, "invalid arguments")
	}

	// todo actually do something
	<-time.After(10 * time.Minute)

	return errors.New("not implemented")
}
