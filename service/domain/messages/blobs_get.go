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

func NewBlobsGet(arguments BlobsGetArguments) (*rpc.Request, error) {
	j, err := arguments.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return rpc.NewRequest(
		BlobsGetProcedure.Name(),
		BlobsGetProcedure.Typ(),
		j,
	)
}

type BlobsGetArguments struct {
	id refs.Blob
}

func NewBlobsGetArguments(id refs.Blob) (BlobsGetArguments, error) {
	if id.IsZero() {
		return BlobsGetArguments{}, errors.New("zero value of id")
	}
	return BlobsGetArguments{id: id}, nil
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

func (a BlobsGetArguments) MarshalJSON() ([]byte, error) {
	args := []string{
		a.id.String(),
	}

	return json.Marshal(args)
}

func (a BlobsGetArguments) Id() refs.Blob {
	return a.id
}
