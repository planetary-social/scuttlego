package messages_test

import (
	"encoding/json"
	"testing"

	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/rooms/features"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestNewRoomMetadata(t *testing.T) {
	req, err := messages.NewRoomMetadata()
	require.NoError(t, err)
	require.Equal(t, rpc.ProcedureTypeAsync, req.Type())
	require.Equal(t, rpc.MustNewProcedureName([]string{"room", "metadata"}), req.Name())
	require.Equal(t, json.RawMessage(`[]`), req.Arguments())
}

func TestNewRoomMetadataResponseFromBytes(t *testing.T) {
	testCases := []struct {
		Name                  string
		String                string
		ExpectedMembership    bool
		ExpectedFeatureTunnel bool
	}{
		{
			Name: "membership_is_true_and_can_tunnel",
			String: `{
"name": "room name",
"membership": true,
"features": ["tunnel", "someunknownfeature"]
}`,
			ExpectedMembership:    true,
			ExpectedFeatureTunnel: true,
		},
		{
			Name: "membership_is_false_and_can_not_tunnel",
			String: `{
"name": "room name",
"membership": false,
"features": ["someunknownfeature"]
}`,
			ExpectedMembership:    false,
			ExpectedFeatureTunnel: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			resp, err := messages.NewRoomMetadataResponseFromBytes([]byte(testCase.String))
			require.NoError(t, err)
			require.Equal(t, testCase.ExpectedMembership, resp.Membership())
			require.Equal(t, testCase.ExpectedFeatureTunnel, resp.Features().Contains(features.FeatureTunnel))
		})
	}
}
