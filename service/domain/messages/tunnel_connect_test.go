package messages_test

import (
	"encoding/json"
	"fmt"
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

func TestNewTunnelConnectToTarget(t *testing.T) {
	portalRef := refs.MustNewIdentity("@gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=.ed25519")
	targetRef := refs.MustNewIdentity("@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519")
	originRef := refs.MustNewIdentity("@650YpEeEBF2H88Z88idG6ZWvWiU2eVG6ov9s1HHEg/E=.ed25519")

	args, err := messages.NewTunnelConnectToTargetArguments(portalRef, targetRef, originRef)
	require.NoError(t, err)

	req, err := messages.NewTunnelConnectToTarget(args)
	require.NoError(t, err)
	require.Equal(t, rpc.ProcedureTypeDuplex, req.Type())
	require.Equal(t, rpc.MustNewProcedureName([]string{"tunnel", "connect"}), req.Name())
	require.Equal(t,
		json.RawMessage(`[{"portal":"@gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=.ed25519","target":"@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519","origin":"@650YpEeEBF2H88Z88idG6ZWvWiU2eVG6ov9s1HHEg/E=.ed25519"}]`),
		req.Arguments(),
	)
}

func TestNewTunnelConnectToTargetArgumentsFromBytes(t *testing.T) {
	portal := refs.MustNewIdentity("@650YpEeEBF2H88Z88idG6ZWvWiU2eVG6ov9s1HHEg/E=.ed25519")
	target := refs.MustNewIdentity("@gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=.ed25519")
	origin := refs.MustNewIdentity("@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519")

	j := fmt.Sprintf(`[{"origin": "%s", "target": "%s", "portal":"%s"}]`, origin, target, portal)

	args, err := messages.NewTunnelConnectToTargetArgumentsFromBytes([]byte(j))
	require.NoError(t, err)
	require.Equal(t, args.Portal(), portal)
	require.Equal(t, args.Target(), target)
	require.Equal(t, args.Origin(), origin)
}
