package debugger_test

//func TestSessionsInitiatedByRemote(t *testing.T) {
//	sessions := debugger.NewSessions()
//
//	msgs := []debugger.Message{
//		{
//			Type:          debugger.MessageTypeReceived,
//			RequestNumber: 1,
//
//			Flags: "some flags",
//			Body:  "some body",
//			Entry: nil,
//		},
//		{
//			Type:          debugger.MessageTypeSent,
//			RequestNumber: -1,
//
//			Flags: "some flags",
//			Body:  "some body",
//			Entry: nil,
//		},
//	}
//
//	for _, msg := range msgs {
//		err := sessions.AddMessage(msg)
//		require.NoError(t, err)
//	}
//
//	require.Equal(t,
//		debugger.Sessions{
//			-1: {
//				Number:      -1,
//				InitiatedBy: debugger.InitiatedByRemoteNode,
//				Messages:    msgs,
//			},
//		},
//		sessions,
//	)
//}
//
//func TestSessionsInitiatedByLocal(t *testing.T) {
//	sessions := debugger.NewSessions()
//
//	msgs := []debugger.Message{
//		{
//			Type:          debugger.MessageTypeSent,
//			RequestNumber: 1,
//
//			Flags: "some flags",
//			Body:  "some body",
//			Entry: nil,
//		},
//		{
//			Type:          debugger.MessageTypeReceived,
//			RequestNumber: -1,
//
//			Flags: "some flags",
//			Body:  "some body",
//			Entry: nil,
//		},
//	}
//
//	for _, msg := range msgs {
//		err := sessions.AddMessage(msg)
//		require.NoError(t, err)
//	}
//
//	require.Equal(t,
//		debugger.Sessions{
//			1: {
//				Number:      1,
//				InitiatedBy: debugger.InitiatedByLocalNode,
//				Messages:    msgs,
//			},
//		},
//		sessions,
//	)
//}
