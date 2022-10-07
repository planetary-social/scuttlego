package messages_test

import (
	"encoding/json"
	"testing"

	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/rooms/aliases"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestNewRoomRevokeAlias(t *testing.T) {
	alias := aliases.MustNewAlias("somealias")

	args, err := messages.NewRoomRevokeAliasArguments(alias)
	require.NoError(t, err)

	req, err := messages.NewRoomRevokeAlias(args)
	require.NoError(t, err)
	require.Equal(t, rpc.ProcedureTypeAsync, req.Type())
	require.Equal(t, rpc.MustNewProcedureName([]string{"room", "revokeAlias"}), req.Name())
	require.Equal(t, json.RawMessage(`["somealias"]`), req.Arguments())
}
