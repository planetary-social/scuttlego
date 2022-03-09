package rpc

import (
	"io"

	rpc2 "github.com/planetary-social/go-ssb/service/domain/network/rpc"
	"github.com/planetary-social/go-ssb/service/domain/network/rpc/messages"

	"github.com/boreq/errors"
)

type BlobStorage interface {
	Get() (io.ReadCloser, error)
}

type HandlerBlobsGet struct {
}

func NewHandlerBlobsGet() *HandlerBlobsGet {
	return &HandlerBlobsGet{}
}

func (h HandlerBlobsGet) Procedure() rpc2.Procedure {
	return messages.BlobsGetProcedure
}

func (h HandlerBlobsGet) Handle(req *rpc2.Request, w *rpc2.ResponseWriter) error {
	_, err := messages.NewBlobsGetArgumentsFromBytes(req.Arguments())
	if err != nil {
		return errors.Wrap(err, "invalid arguments")
	}

	return errors.New("not implemented")
}
