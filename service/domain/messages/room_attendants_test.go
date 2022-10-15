package messages_test

import (
	"encoding/json"
	"testing"

	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestNewRoomAttendants(t *testing.T) {
	req, err := messages.NewRoomAttendants()
	require.NoError(t, err)
	require.Equal(t, rpc.ProcedureTypeSource, req.Type())
	require.Equal(t, rpc.MustNewProcedureName([]string{"room", "attendants"}), req.Name())
	require.Equal(t, json.RawMessage("[]"), req.Arguments())
}

func TestNewRoomAttendantsResponseStateFromBytes(t *testing.T) {
	s := `
{
  "type":"state",
  "ids": [
    "@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519",
    "@gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=.ed25519"
  ]
}
`
	resp, err := messages.NewRoomAttendantsResponseStateFromBytes([]byte(s))
	require.NoError(t, err)

	require.Equal(t,
		[]refs.Identity{
			refs.MustNewIdentity("@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519"),
			refs.MustNewIdentity("@gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=.ed25519"),
		},
		resp.Ids(),
	)
}

func TestNewRoomAttendantsResponseJoinedOrLeftFromBytes(t *testing.T) {
	testCases := []struct {
		Name        string
		String      string
		ExpectedTyp messages.RoomAttendantsResponseType
		ExpectedId  refs.Identity
	}{
		{
			Name: "joined",
			String: `
{
  "type":"joined",
  "id": "@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519"
}
`,
			ExpectedTyp: messages.RoomAttendantsResponseTypeJoined,
			ExpectedId:  refs.MustNewIdentity("@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519"),
		},
		{
			Name: "left",
			String: `
{
  "type":"left",
  "id": "@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519"
}
`,
			ExpectedTyp: messages.RoomAttendantsResponseTypeLeft,
			ExpectedId:  refs.MustNewIdentity("@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			resp, err := messages.NewRoomAttendantsResponseJoinedOrLeftFromBytes([]byte(testCase.String))
			require.NoError(t, err)
			require.Equal(t, testCase.ExpectedTyp, resp.Typ())
			require.Equal(t, testCase.ExpectedId, resp.Id())
		})
	}
}
