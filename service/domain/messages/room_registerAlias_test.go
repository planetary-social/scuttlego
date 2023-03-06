package messages_test

import (
	"encoding/base64"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/rooms/aliases"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestNewRoomRegisterAlias(t *testing.T) {
	alias := aliases.MustNewAlias("somealias")
	userIdentity := fixtures.SomePrivateIdentity()
	userRef := refs.MustNewIdentityFromPublic(userIdentity.Public())
	roomRef := fixtures.SomeRefIdentity()

	message, err := aliases.NewRegistrationMessage(alias, userRef, roomRef)
	require.NoError(t, err)

	signature, err := aliases.NewRegistrationSignature(message, userIdentity)
	require.NoError(t, err)

	args, err := messages.NewRoomRegisterAliasArguments(alias, signature)
	require.NoError(t, err)

	req, err := messages.NewRoomRegisterAlias(args)
	require.NoError(t, err)
	require.Equal(t, rpc.ProcedureTypeAsync, req.Type())
	require.Equal(t, rpc.MustNewProcedureName([]string{"room", "registerAlias"}), req.Name())

	var actualArgs []string
	err = jsoniter.Unmarshal(req.Arguments(), &actualArgs)
	require.NoError(t, err)
	require.Equal(t,
		[]string{
			alias.String(),
			base64.StdEncoding.EncodeToString(signature.Bytes()) + ".sig.ed25519",
		},
		actualArgs,
	)
}

func TestNewRoomRegisterAliasResponseFromBytes(t *testing.T) {
	resp, err := messages.NewRoomRegisterAliasResponseFromBytes([]byte("somealiasurl"))
	require.NoError(t, err)
	require.Equal(t,
		aliases.MustNewAliasEndpointURL("somealiasurl"),
		resp.AliasEndpointURL(),
	)
}
