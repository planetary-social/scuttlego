package tunnel_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/rooms/tunnel"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDialer_DialViaRoomPerformsCorrectRequestsAndCallsInitializerWithCorrectArguments(t *testing.T) {
	clientPeerInitializer := newClientPeerInitializerMock()
	dialer := tunnel.NewDialer(clientPeerInitializer)

	ctx := fixtures.TestContext(t)
	portalConn := mocks.NewConnectionMock(ctx)
	portalPeerRef := fixtures.SomeRefIdentity()
	portalPeer := transport.MustNewPeer(
		portalPeerRef.Identity(),
		portalConn,
	)
	targetRef := fixtures.SomeRefIdentity()

	someResponseBytes := fixtures.SomeBytes()

	portalConn.Mock(func(req *rpc.Request) []rpc.ResponseWithError {
		assert.Equal(t, messages.TunnelConnectProcedure.Name(), req.Name())
		assert.Equal(t, messages.TunnelConnectProcedure.Typ(), req.Type())
		assert.Equal(t,
			fmt.Sprintf(`[{"portal":"%s","target":"%s"}]`, portalPeerRef.String(), targetRef.String()),
			string(req.Arguments()),
		)

		return []rpc.ResponseWithError{
			{
				Value: rpc.NewResponse(someResponseBytes),
				Err:   nil,
			},
		}
	})

	_, err := dialer.DialViaRoom(ctx, portalPeer, targetRef.Identity())
	require.NoError(t, err)

	require.Eventually(t,
		func() bool {
			if len(clientPeerInitializer.calls) > 0 {
				fmt.Println(targetRef.Identity().String(), clientPeerInitializer.calls[0].Remote.String())
				fmt.Println(len(someResponseBytes), len(clientPeerInitializer.calls[0].ReceivedMessage))
			}
			return assert.ObjectsAreEqual(clientPeerInitializer.calls, []clientPeerInitializerCall{
				{
					Remote:          targetRef.Identity(),
					ReceivedMessage: someResponseBytes,
				},
			})

		}, 1*time.Second, 10*time.Millisecond)
}

type clientPeerInitializerMock struct {
	calls []clientPeerInitializerCall
}

func newClientPeerInitializerMock() *clientPeerInitializerMock {
	return &clientPeerInitializerMock{}
}

func (c *clientPeerInitializerMock) InitializeClientPeer(ctx context.Context, rwc io.ReadWriteCloser, remote identity.Public) (transport.Peer, error) {
	buf := make([]byte, 10000)
	n, err := rwc.Read(buf)
	if err != nil {
		return transport.Peer{}, err
	}

	c.calls = append(c.calls, clientPeerInitializerCall{
		ReceivedMessage: buf[:n],
		Remote:          remote,
	})

	return transport.Peer{}, nil
}

type clientPeerInitializerCall struct {
	Remote          identity.Public
	ReceivedMessage []byte
}
