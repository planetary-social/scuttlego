package messages

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/network/rpc"
)

var (
	BlobsGetProcedureName = rpc.MustNewProcedureName([]string{"blobs", "get"})
	BlobsGetProcedure     = rpc.MustNewProcedure(BlobsGetProcedureName, rpc.ProcedureTypeSource)
)

type BlobsGetArguments struct {
	Id string // todo blob id
}

func NewBlobsGetArgumentsFromBytes(b []byte) (BlobsGetArguments, error) {
	var args []string

	if err := json.Unmarshal(b, &args); err != nil {
		return BlobsGetArguments{}, errors.Wrap(err, "json unmarshal failed")
	}

	if len(args) != 1 {
		return BlobsGetArguments{}, errors.New("expected exactly one argument")
	}

	// todo validate argument by using a constructor
	return BlobsGetArguments{
		Id: args[0],
	}, nil
}
