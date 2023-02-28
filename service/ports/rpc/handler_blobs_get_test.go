package rpc_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	transportrpc "github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux/mocks"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/planetary-social/scuttlego/service/ports/rpc"
	"github.com/stretchr/testify/require"
)

func TestArgumentsArePassedToQuery(t *testing.T) {
	hash := fixtures.SomeRefBlob()
	data := fixtures.SomeBytes()

	testCases := []struct {
		Name string

		Hash refs.Blob
		Size *blobs.Size
		Max  *blobs.Size

		ExpectedQuery queries.GetBlob
	}{
		{
			Name: "hash",

			Hash: hash,
			Size: nil,
			Max:  nil,

			ExpectedQuery: queries.GetBlob{
				Id:   hash,
				Size: nil,
				Max:  nil,
			},
		},
		{
			Name: "hash_and_size",

			Hash: hash,
			Size: internal.Ptr(blobs.MustNewSize(int64(len(data)))),
			Max:  nil,

			ExpectedQuery: queries.GetBlob{
				Id:   hash,
				Size: internal.Ptr(blobs.MustNewSize(int64(len(data)))),
				Max:  nil,
			},
		},
		{
			Name: "hash_and_max",

			Hash: hash,
			Size: nil,
			Max:  internal.Ptr(blobs.MustNewSize(int64(len(data)))),

			ExpectedQuery: queries.GetBlob{
				Id:   hash,
				Size: nil,
				Max:  internal.Ptr(blobs.MustNewSize(int64(len(data)))),
			},
		},
		{
			Name: "hash_and_size_and_max",

			Hash: hash,
			Size: internal.Ptr(blobs.MustNewSize(int64(len(data)))),
			Max:  internal.Ptr(blobs.MustNewSize(int64(len(data)))),

			ExpectedQuery: queries.GetBlob{
				Id:   hash,
				Size: internal.Ptr(blobs.MustNewSize(int64(len(data)))),
				Max:  internal.Ptr(blobs.MustNewSize(int64(len(data)))),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			queryHandler := newGetBlobQueryHandlerMock()
			queryHandler.MockBlob(hash, data)

			h := rpc.NewHandlerBlobsGet(queryHandler)

			ctx := fixtures.TestContext(t)
			s := mocks.NewMockCloserStream()
			req := createBlobsGetRequest(t, testCase.Hash, testCase.Size, testCase.Max)

			err := h.Handle(ctx, s, req)
			require.NoError(t, err)

			require.Equal(t, []queries.GetBlob{testCase.ExpectedQuery}, queryHandler.Calls)
		})
	}
}

func TestIfHandlerReturnsErrorNoMessagesAreSent(t *testing.T) {
	queryHandler := newGetBlobQueryHandlerMock()
	h := rpc.NewHandlerBlobsGet(queryHandler)

	ctx := fixtures.TestContext(t)
	s := mocks.NewMockCloserStream()
	req := createBlobsGetRequest(t, fixtures.SomeRefBlob(), nil, nil)

	err := h.Handle(ctx, s, req)
	require.Error(t, err)

	require.Empty(t, s.WrittenMessages())
}

func TestSmallBlobIsWrittenToResponseWriter(t *testing.T) {
	queryHandler := newGetBlobQueryHandlerMock()
	h := rpc.NewHandlerBlobsGet(queryHandler)

	mockData := []byte("some-fake-blob-data")

	id := fixtures.SomeRefBlob()
	queryHandler.MockBlob(id, mockData)

	ctx := fixtures.TestContext(t)
	s := mocks.NewMockCloserStream()
	req := createBlobsGetRequest(t, id, nil, nil)

	err := h.Handle(ctx, s, req)
	require.NoError(t, err)

	require.Equal(t,
		[]mocks.MockCloserStreamWriteMessageCall{
			{
				Body:     mockData,
				BodyType: transport.MessageBodyTypeBinary,
			},
		},
		s.WrittenMessages(),
	)

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
	s := mocks.NewMockCloserStream()
	req := createBlobsGetRequest(t, id, nil, nil)

	err := h.Handle(ctx, s, req)
	require.NoError(t, err)

	require.Equal(t,
		[]mocks.MockCloserStreamWriteMessageCall{
			{
				Body:     payloadInFirstMessage,
				BodyType: transport.MessageBodyTypeBinary,
			},
			{
				Body:     payloadInSecondMessage,
				BodyType: transport.MessageBodyTypeBinary,
			},
			{
				Body:     payloadInThirdMessage,
				BodyType: transport.MessageBodyTypeBinary,
			},
		},
		s.WrittenMessages(),
	)

	queryHandler.RequireThereAreNoOpenReadClosers(t)
}

type getBlobQueryHandlerMock struct {
	Calls           []queries.GetBlob
	blobs           map[string][]byte
	openReadClosers map[*readCloserTrackingCloses]struct{}
}

func newGetBlobQueryHandlerMock() *getBlobQueryHandlerMock {
	return &getBlobQueryHandlerMock{
		blobs:           make(map[string][]byte),
		openReadClosers: make(map[*readCloserTrackingCloses]struct{}),
	}
}

func (h *getBlobQueryHandlerMock) Handle(query queries.GetBlob) (io.ReadCloser, error) {
	h.Calls = append(h.Calls, query)

	data, ok := h.blobs[query.Id.String()]
	if !ok {
		return nil, errors.New("blob not found")
	}
	rc := newReadCloserTrackingCloses(bytes.NewBuffer(data), h.onReadCloserClose)
	h.openReadClosers[rc] = struct{}{}
	return rc, nil
}

func (h *getBlobQueryHandlerMock) onReadCloserClose(rc *readCloserTrackingCloses) {
	delete(h.openReadClosers, rc)
}

func (h getBlobQueryHandlerMock) MockBlob(id refs.Blob, data []byte) {
	cpy := make([]byte, len(data))
	copy(cpy, data)
	h.blobs[id.String()] = cpy
}

func (h getBlobQueryHandlerMock) RequireThereAreNoOpenReadClosers(t *testing.T) {
	require.Empty(t, h.openReadClosers)
}

func createBlobsGetRequest(t *testing.T, id refs.Blob, size, max *blobs.Size) *transportrpc.Request {
	args, err := messages.NewBlobsGetArguments(id, size, max)
	require.NoError(t, err)

	req, err := messages.NewBlobsGet(args)
	require.NoError(t, err)

	return req
}

type onCloseFn func(*readCloserTrackingCloses)

type readCloserTrackingCloses struct {
	reader  io.Reader
	onClose onCloseFn
}

func newReadCloserTrackingCloses(reader io.Reader, onClose onCloseFn) *readCloserTrackingCloses {
	rc := &readCloserTrackingCloses{
		reader:  reader,
		onClose: onClose,
	}
	return rc
}

func (r *readCloserTrackingCloses) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *readCloserTrackingCloses) Close() error {
	r.onClose(r)
	return nil
}
