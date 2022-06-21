package blobs_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/adapters/blobs"
	blobsdomain "github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	directory := fixtures.Directory(t)
	logger := logging.NewDevNullLogger()

	storage, err := blobs.NewFilesystemStorage(directory, logger)
	require.NoError(t, err)

	id, r := newFakeBlob(t)

	err = storage.Store(id, r)
	require.NoError(t, err)
}

func newFakeBlob(t *testing.T) (refs.Blob, io.Reader) {
	buf := &bytes.Buffer{}

	h := blobsdomain.NewHasher()
	w := io.MultiWriter(h, buf)

	data := fixtures.SomeBytes()
	_, err := w.Write(data)
	require.NoError(t, err)

	id, err := h.SumRef()
	require.NoError(t, err)

	return id, buf
}
