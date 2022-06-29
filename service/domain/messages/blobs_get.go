package messages

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
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
	hash      refs.Blob
	size, max *blobs.Size
}

func NewBlobsGetArguments(hash refs.Blob, size, max *blobs.Size) (BlobsGetArguments, error) {
	if hash.IsZero() {
		return BlobsGetArguments{}, errors.New("zero value of hash")
	}

	if size != nil && size.IsZero() {
		return BlobsGetArguments{}, errors.New("size can't be zero")
	}

	if max != nil && max.IsZero() {
		return BlobsGetArguments{}, errors.New("max can't be zero")
	}

	return BlobsGetArguments{
		hash: hash,
		size: size,
		max:  max,
	}, nil
}

func NewBlobsGetArgumentsFromBytes(b []byte) (BlobsGetArguments, error) {
	var err error

	args, stringErr := newBlobsGetArgumentsFromBytesString(b)
	err = multierror.Append(err, stringErr)
	if stringErr == nil {
		return args, nil
	}

	args, objectErr := newBlobsGetArgumentsFromBytesObject(b)
	err = multierror.Append(err, objectErr)
	if objectErr == nil {
		return args, nil
	}

	return BlobsGetArguments{}, err
}

func newBlobsGetArgumentsFromBytesString(b []byte) (BlobsGetArguments, error) {
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

	return NewBlobsGetArguments(id, nil, nil)
}

func newBlobsGetArgumentsFromBytesObject(b []byte) (BlobsGetArguments, error) {
	var args []blobsGetArgumentsTransport

	if err := json.Unmarshal(b, &args); err != nil {
		return BlobsGetArguments{}, errors.Wrap(err, "json unmarshal failed")
	}

	if len(args) != 1 {
		return BlobsGetArguments{}, errors.New("expected exactly one argument")
	}

	id, err := refs.NewBlob(args[0].Hash)
	if err != nil {
		return BlobsGetArguments{}, errors.Wrap(err, "could not create a blob ref")
	}

	size, err := sizeOrNil(args[0].Size)
	if err != nil {
		return BlobsGetArguments{}, errors.Wrap(err, "failed to parse size")
	}

	max, err := sizeOrNil(args[0].Max)
	if err != nil {
		return BlobsGetArguments{}, errors.Wrap(err, "failed to parse max")
	}

	return NewBlobsGetArguments(id, size, max)
}

func (a BlobsGetArguments) MarshalJSON() ([]byte, error) {
	if a.size == nil && a.max == nil {
		args := []string{
			a.hash.String(),
		}

		return json.Marshal(args)
	}

	args := []blobsGetArgumentsTransport{
		{
			Hash: a.hash.String(),
		},
	}

	if a.size != nil {
		v := a.size.InBytes()
		args[0].Size = &v
	}

	if a.max != nil {
		v := a.max.InBytes()
		args[0].Max = &v
	}

	return json.Marshal(args)
}

func (a BlobsGetArguments) Hash() refs.Blob {
	return a.hash
}

func (a BlobsGetArguments) Size() (blobs.Size, bool) {
	if a.size != nil {
		return *a.size, true
	}
	return blobs.Size{}, false
}

func (a BlobsGetArguments) Max() (blobs.Size, bool) {
	if a.max != nil {
		return *a.max, true
	}
	return blobs.Size{}, false
}

type blobsGetArgumentsTransport struct {
	Hash string `json:"hash"`
	Size *int64 `json:"size,omitempty"`
	Max  *int64 `json:"max,omitempty"`
}

func sizeOrNil(n *int64) (*blobs.Size, error) {
	if n == nil {
		return nil, nil
	}

	size, err := blobs.NewSize(*n)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a size")
	}

	return &size, nil
}
