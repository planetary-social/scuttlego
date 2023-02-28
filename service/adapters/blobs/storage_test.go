package blobs_test

import (
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters/blobs"
	blobsdomain "github.com/planetary-social/scuttlego/service/domain/blobs"
	blobReplication "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestStorageStore(t *testing.T) {
	directory := fixtures.Directory(t)
	logger := logging.NewDevNullLogger()

	storage, err := blobs.NewFilesystemStorage(directory, logger)
	require.NoError(t, err)

	id, r, data := newFakeBlob(t)

	err = storage.Store(id, r)
	require.NoError(t, err)

	size, err := storage.Size(id)
	require.NoError(t, err)
	require.EqualValues(t, len(data), size.InBytes())

	rc, err := storage.Get(id)
	require.NoError(t, err)
	defer rc.Close()

	readData, err := io.ReadAll(rc)
	require.NoError(t, err)

	require.Equal(t, data, readData)
}

func TestSizeReturnsBlobNotFound(t *testing.T) {
	directory := fixtures.Directory(t)
	logger := logging.NewDevNullLogger()

	storage, err := blobs.NewFilesystemStorage(directory, logger)
	require.NoError(t, err)

	_, err = storage.Size(fixtures.SomeRefBlob())
	require.ErrorIs(t, err, blobReplication.ErrBlobNotFound)
}

func TestStorageCreate(t *testing.T) {
	directory := fixtures.Directory(t)
	logger := logging.NewDevNullLogger()

	storage, err := blobs.NewFilesystemStorage(directory, logger)
	require.NoError(t, err)

	bts := fixtures.SomeBytes()

	id, err := storage.Create(bytes.NewReader(bts))
	require.NoError(t, err)
	require.NotEmpty(t, id.String())

	size, err := storage.Size(id)
	require.NoError(t, err)
	require.EqualValues(t, len(bts), size.InBytes())

	rc, err := storage.Get(id)
	require.NoError(t, err)
	defer rc.Close()

	readData, err := io.ReadAll(rc)
	require.NoError(t, err)

	require.Equal(t, bts, readData)
}

func TestStorageStoreUsesSamePathsAsGoSSB(t *testing.T) {
	directory := fixtures.Directory(t)
	logger := logging.NewDevNullLogger()

	storage, err := blobs.NewFilesystemStorage(directory, logger)
	require.NoError(t, err)

	data := []byte("testblobdata")
	id, r := newBlobFromData(t, data)

	err = storage.Store(id, r)
	require.NoError(t, err)

	var pathsInDirectory []string
	err = filepath.Walk(directory, func(path string, info fs.FileInfo, err error) error {
		pathsInDirectory = append(pathsInDirectory, path)
		return nil
	})
	require.NoError(t, err)

	require.Equal(t,
		[]string{
			directory,
			filepath.Join(directory, "sha256"),
			filepath.Join(directory, "sha256/5c"),
			filepath.Join(directory, "sha256/5c/dfa36516194ff23d296063290032cb65d2e649797bbf3966c8813270aaaaf1"),
			filepath.Join(directory, "tmp"),
		},
		pathsInDirectory,
	)
}

func newFakeBlob(t *testing.T) (refs.Blob, io.Reader, []byte) {
	data := fixtures.SomeBytes()
	ref, reader := newBlobFromData(t, data)
	return ref, reader, data
}

func newBlobFromData(t *testing.T, data []byte) (refs.Blob, io.Reader) {
	h := blobsdomain.NewHasher()
	_, err := h.Write(data)
	require.NoError(t, err)

	id, err := h.SumRef()
	require.NoError(t, err)

	return id, bytes.NewReader(data)
}
