package rpc_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	transportrpc "github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	"github.com/planetary-social/go-ssb/service/ports/rpc"
	"github.com/stretchr/testify/require"
)

func TestIfHandlerReturnsErrorNoMessagesAreSent(t *testing.T) {
	queryHandler := newGetBlobQueryHandlerMock()
	h := rpc.NewHandlerBlobsGet(queryHandler)

	ctx := fixtures.TestContext(t)
	rw := NewMockResponseWriter()
	req := createBlobsGetRequest(t, fixtures.SomeRefBlob())

	err := h.Handle(ctx, rw, req)
	require.Error(t, err)

	require.Empty(t, rw.WrittenMessages)
}

func TestSmallBlobIsWrittenToResponseWriter(t *testing.T) {
	queryHandler := newGetBlobQueryHandlerMock()
	h := rpc.NewHandlerBlobsGet(queryHandler)

	id := fixtures.SomeRefBlob()
	queryHandler.MockBlob(id, []byte("test"))

	ctx := fixtures.TestContext(t)
	rw := NewMockResponseWriter()
	req := createBlobsGetRequest(t, id)

	err := h.Handle(ctx, rw, req)
	require.NoError(t, err)

	require.Len(t, rw.WrittenMessages, 1)
	require.Equal(t, []byte("test"), rw.WrittenMessages[0])

	queryHandler.RequireThereAreNoOpenReadClosers(t)
}

func TestLargeBlobIsWrittenToResponseWriter(t *testing.T) {
	queryHandler := newGetBlobQueryHandlerMock()
	h := rpc.NewHandlerBlobsGet(queryHandler)

	payloadInFirstMessage := []byte(strings.Repeat("a", rpc.MaxBlobChunkSizeInBytes))
	payloadInSecondMessage := []byte(strings.Repeat("b", rpc.MaxBlobChunkSizeInBytes))
	payloadInThirdMessage := []byte("something that doesn't fill up the message completely")

	var payload []byte
	payload = append(payload, payloadInFirstMessage...)
	payload = append(payload, payloadInSecondMessage...)
	payload = append(payload, payloadInThirdMessage...)

	id := fixtures.SomeRefBlob()
	queryHandler.MockBlob(id, payload)

	ctx := fixtures.TestContext(t)
	rw := NewMockResponseWriter()
	req := createBlobsGetRequest(t, id)

	err := h.Handle(ctx, rw, req)
	require.NoError(t, err)

	require.Len(t, rw.WrittenMessages, 3)
	require.Equal(t, payloadInFirstMessage, rw.WrittenMessages[0])
	require.Equal(t, payloadInSecondMessage, rw.WrittenMessages[1])
	require.Equal(t, payloadInThirdMessage, rw.WrittenMessages[2])

	queryHandler.RequireThereAreNoOpenReadClosers(t)
}

type getBlobQueryHandlerMock struct {
	blobs           map[string][]byte
	openReadClosers map[*readCloserTrackingCloses]struct{}
}

func newGetBlobQueryHandlerMock() *getBlobQueryHandlerMock {
	return &getBlobQueryHandlerMock{
		blobs:           make(map[string][]byte),
		openReadClosers: make(map[*readCloserTrackingCloses]struct{}),
	}
}

func (h getBlobQueryHandlerMock) Handle(query queries.GetBlob) (io.ReadCloser, error) {
	data, ok := h.blobs[query.Id.String()]
	if !ok {
		return nil, errors.New("blob not found")
	}
	return newReadCloserTrackingCloses(bytes.NewBuffer(data), h.openReadClosers), nil
}

func (h getBlobQueryHandlerMock) MockBlob(id refs.Blob, data []byte) {
	cpy := make([]byte, len(data))
	copy(cpy, data)
	h.blobs[id.String()] = cpy
}

func (h getBlobQueryHandlerMock) RequireThereAreNoOpenReadClosers(t *testing.T) {
	require.Empty(t, h.openReadClosers)
}

func createBlobsGetRequest(t *testing.T, id refs.Blob) *transportrpc.Request {
	args, err := messages.NewBlobsGetArguments(
		id,
	)
	require.NoError(t, err)

	argsBytes, err := args.MarshalJSON()
	require.NoError(t, err)

	req, err := transportrpc.NewRequest(
		fixtures.SomeProcedureName(),
		fixtures.SomeProcedureType(),
		argsBytes,
	)
	require.NoError(t, err)

	return req
}

type readCloserTrackingCloses struct {
	reader io.Reader
	m      map[*readCloserTrackingCloses]struct{}
}

func newReadCloserTrackingCloses(reader io.Reader, m map[*readCloserTrackingCloses]struct{}) *readCloserTrackingCloses {
	rc := &readCloserTrackingCloses{
		reader: reader,
		m:      m,
	}
	m[rc] = struct{}{}
	return rc
}

func (r *readCloserTrackingCloses) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *readCloserTrackingCloses) Close() error {
	delete(r.m, r)
	return nil
}
