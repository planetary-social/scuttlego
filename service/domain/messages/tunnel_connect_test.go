package messages_test

import (
	"encoding/json"
	"testing"

	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestNewTunnelConnectToPortal(t *testing.T) {
	portalRef := refs.MustNewIdentity("@gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=.ed25519")
	targetRef := refs.MustNewIdentity("@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519")

	args, err := messages.NewTunnelConnectToPortalArguments(portalRef, targetRef)
	require.NoError(t, err)

	req, err := messages.NewTunnelConnectToPortal(args)
	require.NoError(t, err)
	require.Equal(t, rpc.ProcedureTypeDuplex, req.Type())
	require.Equal(t, rpc.MustNewProcedureName([]string{"tunnel", "connect"}), req.Name())
	require.Equal(t,
		json.RawMessage(`[{"portal":"@gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=.ed25519","target":"@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519"}]`),
		req.Arguments(),
	)
}
