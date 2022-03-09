package rpc

import (
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/network/rpc"
	"github.com/planetary-social/go-ssb/network/rpc/messages"
)

type BlobStorage interface {
	Get() (io.ReadCloser, error)
}

type HandlerBlobsGet struct {
}

func NewHandlerBlobsGet() *HandlerBlobsGet {
	return &HandlerBlobsGet{}
}

func (h HandlerBlobsGet) Procedure() rpc.Procedure {
	return messages.BlobsGetProcedure
}

func (h HandlerBlobsGet) Handle(req *rpc.Request, w *rpc.ResponseWriter) error {
	_, err := messages.NewBlobsGetArgumentsFromBytes(req.Arguments())
	if err != nil {
		return errors.Wrap(err, "invalid arguments")
	}

	return errors.New("not implemented")
}
