package messages

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

var (
	BlobsGetProcedureName = rpc.MustNewProcedureName([]string{"blobs", "get"})
	BlobsGetProcedure     = rpc.MustNewProcedure(BlobsGetProcedureName, rpc.ProcedureTypeSource)
)

type BlobsGetArguments struct {
	id refs.Blob
}

func NewBlobsGetArgumentsFromBytes(b []byte) (BlobsGetArguments, error) {
	var args []string

	if err := json.Unmarshal(b, &args); err != nil {
		return BlobsGetArguments{}, errors.Wrap(err, "json unmarshal failed")
	}

	if len(args) != 1 {
		return BlobsGetArguments{}, errors.New("expected exactly one argument")
	}

	id, err := refs.NewBlob(args[0])
	if err != nil {
		return BlobsGetArguments{}, errors.Wrap(err, "could not create a blob ref")
	}

	return NewBlobsGetArguments(id)
}

func NewBlobsGetArguments(id refs.Blob) (BlobsGetArguments, error) {
	if id.IsZero() {
		return BlobsGetArguments{}, errors.New("zero value of blob id")
	}

	return BlobsGetArguments{
		id: id,
	}, nil
}

func (b BlobsGetArguments) Id() refs.Blob {
	return b.id
}
