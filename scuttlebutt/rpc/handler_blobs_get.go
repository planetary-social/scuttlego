package rpc

import (
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/network/rpc"
	rpc2 "github.com/planetary-social/go-ssb/network/rpc/messages"
)

type BlobStorage interface {
	Get() (io.ReadCloser, error)
}

type HandlerBlobsGet struct {
}

func (h HandlerBlobsGet) Procedure() rpc.Procedure {
	return rpc2.BlobsGetProcedure
}

func (h HandlerBlobsGet) Handle(req rpc.Request, w rpc.ResponseWriter) error {
	_, err := rpc2.NewBlobsGetArgumentsFromBytes(req.Arguments())
	if err != nil {
		return errors.Wrap(err, "invalid arguments")
	}

	return errors.New("not implemented")
}
