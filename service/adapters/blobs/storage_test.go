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

	id, r, data := newFakeBlob(t)

	err = storage.Store(id, r)
	require.NoError(t, err)

	rc, err := storage.GetBlob(id)
	require.NoError(t, err)
	defer rc.Close()

	readData, err := io.ReadAll(rc)
	require.NoError(t, err)

	require.Equal(t, data, readData)
}

func newFakeBlob(t *testing.T) (refs.Blob, io.Reader, []byte) {
	buf := &bytes.Buffer{}

	h := blobsdomain.NewHasher()
	w := io.MultiWriter(h, buf)

	data := fixtures.SomeBytes()
	_, err := w.Write(data)
	require.NoError(t, err)

	id, err := h.SumRef()
	require.NoError(t, err)

	return id, buf, data
}
