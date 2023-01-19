package debugger_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/cmd/log-debugger/debugger"
	"github.com/stretchr/testify/require"
)

func TestSessionsInitiatedByRemote(t *testing.T) {
	sessions := debugger.NewSessions()

	msgs := []debugger.Message{
		{
			Type:          debugger.MessageTypeReceived,
			RequestNumber: 1,

			Flags: "some flags",
			Body:  "some body",
			Entry: nil,
		},
		{
			Type:          debugger.MessageTypeSent,
			RequestNumber: -1,

			Flags: "some flags",
			Body:  "some body",
			Entry: nil,
		},
	}

	for _, msg := range msgs {
		err := sessions.AddMessage(msg)
		require.NoError(t, err)
	}

	require.Equal(t,
		debugger.Sessions{
			-1: {
				Number:       -1,
				InititatedBy: debugger.InitiatedByRemoteNode,
				Messages:     msgs,
			},
		},
		sessions,
	)
}

func TestSessionsInitiatedByLocal(t *testing.T) {
	sessions := debugger.NewSessions()

	msgs := []debugger.Message{
		{
			Type:          debugger.MessageTypeSent,
			RequestNumber: 1,

			Flags: "some flags",
			Body:  "some body",
			Entry: nil,
		},
		{
			Type:          debugger.MessageTypeReceived,
			RequestNumber: -1,

			Flags: "some flags",
			Body:  "some body",
			Entry: nil,
		},
	}

	for _, msg := range msgs {
		err := sessions.AddMessage(msg)
		require.NoError(t, err)
	}

	require.Equal(t,
		debugger.Sessions{
			1: {
				Number:       1,
				InititatedBy: debugger.InitiatedByLocalNode,
				Messages:     msgs,
			},
		},
		sessions,
	)
}
