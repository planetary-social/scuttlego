package messages_test

import (
	"encoding/json"
	"testing"

	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/rooms/aliases"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestNewRoomListAliases(t *testing.T) {
	identityRef := refs.MustNewIdentity("@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519")

	args, err := messages.NewRoomListAliasesArguments(identityRef)
	require.NoError(t, err)

	req, err := messages.NewRoomListAliases(args)
	require.NoError(t, err)
	require.Equal(t, rpc.ProcedureTypeAsync, req.Type())
	require.Equal(t, rpc.MustNewProcedureName([]string{"room", "listAliases"}), req.Name())
	require.Equal(t, json.RawMessage(`["@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519"]`), req.Arguments())
}

func TestNewRoomListAliasesResponseFromBytes(t *testing.T) {
	s := `["alias1", "alias2"]`
	resp, err := messages.NewRoomListAliasesResponseFromBytes([]byte(s))
	require.NoError(t, err)
	require.Equal(t,
		[]aliases.Alias{
			aliases.MustNewAlias("alias1"),
			aliases.MustNewAlias("alias2"),
		},
		resp.Aliases(),
	)
}
