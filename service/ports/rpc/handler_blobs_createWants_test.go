package rpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	transportrpc "github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux/mocks"
	"github.com/planetary-social/scuttlego/service/ports/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWantsHandlerSendsValuesReturnedByCreateWantsCommand(t *testing.T) {
	commandHandler := newCreateWantsCommandHandlerMock()

	blob1 := refs.MustNewBlob("&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256")
	blob2 := refs.MustNewBlob("&gYVaHgAWeTnLZpTSxCKs0gigByk5SH9pmeudGKRHhAQ=.sha256")
	size := blobs.MustNewSize(123)
	distance := blobs.MustNewWantDistance(1)

	commandHandler.BlobsToReturn = []messages.BlobWithSizeOrWantDistance{
		messages.MustNewBlobWithSize(blob1, size),
		messages.MustNewBlobWithWantDistance(blob2, distance),
	}

	h := rpc.NewHandlerBlobsCreateWants(commandHandler)

	ctx := fixtures.TestContext(t)
	s := mocks.NewMockCloserStream()
	req := createWantsRequest(t)

	err := h.Handle(ctx, s, req)
	require.NoError(t, err)

	require.Eventually(t,
		func() bool {
			for i, msg := range s.WrittenMessages() {
				t.Log(i, string(msg))
			}
			return assert.ObjectsAreEqual(
				[][]byte{
					[]byte(`{"\u0026Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256":123}`),
					[]byte(`{"\u0026gYVaHgAWeTnLZpTSxCKs0gigByk5SH9pmeudGKRHhAQ=.sha256":-1}`),
				},
				s.WrittenMessages(),
			)
		},
		1*time.Second,
		10*time.Millisecond,
	)
}

type createWantsCommandHandlerMock struct {
	BlobsToReturn []messages.BlobWithSizeOrWantDistance
}

func newCreateWantsCommandHandlerMock() *createWantsCommandHandlerMock {
	return &createWantsCommandHandlerMock{}
}

func (c createWantsCommandHandlerMock) Handle(ctx context.Context, cmd commands.CreateWants) (<-chan messages.BlobWithSizeOrWantDistance, error) {
	ch := make(chan messages.BlobWithSizeOrWantDistance)

	go func() {
		defer close(ch)

		for _, v := range c.BlobsToReturn {
			select {
			case ch <- v:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func createWantsRequest(t *testing.T) *transportrpc.Request {
	req, err := messages.NewBlobsCreateWants()
	require.NoError(t, err)
	return req
}
