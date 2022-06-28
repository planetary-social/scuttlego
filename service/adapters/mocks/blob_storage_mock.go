package mocks

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type BlobStorageMock struct {
	blobs map[string][]byte
}

func NewBlobStorageMock() *BlobStorageMock {
	return &BlobStorageMock{
		blobs: make(map[string][]byte),
	}
}

func (b BlobStorageMock) GetBlob(id refs.Blob) (io.ReadCloser, error) {
	data, ok := b.blobs[id.String()]
	if !ok {
		return nil, errors.New("blob not found")
	}
	return ioutil.NopCloser(bytes.NewBuffer(data)), nil
}

func (b BlobStorageMock) MockBlob(id refs.Blob, data []byte) {
	cpy := make([]byte, len(data))
	copy(cpy, data)
	b.blobs[id.String()] = cpy
}
