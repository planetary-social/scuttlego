package transport_test

import (
	"testing"

	msgcontents "github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestMappingContactUnmarshal(t *testing.T) {
	makeContactWithActions := func(actions []msgcontents.ContactAction) msgcontents.Contact {
		return msgcontents.MustNewContact(
			refs.MustNewIdentity("@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519"),
			msgcontents.MustNewContactActions(actions),
		)
	}

	testCases := []struct {
		Name            string
		Content         string
		ExpectedMessage msgcontents.KnownMessageContent
	}{
		{
			Name: "missing_action",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519"
}`,
			ExpectedMessage: nil,
		},
		{
			Name: "following",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"following": true
}`,
			ExpectedMessage: makeContactWithActions([]msgcontents.ContactAction{
				msgcontents.ContactActionFollow,
			}),
		},
		{
			Name: "unfollowing",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"following": false
}`,
			ExpectedMessage: makeContactWithActions([]msgcontents.ContactAction{
				msgcontents.ContactActionUnfollow,
			}),
		},
		{
			Name: "blocking",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"blocking": true
}`,
			ExpectedMessage: makeContactWithActions([]msgcontents.ContactAction{
				msgcontents.ContactActionBlock,
			}),
		},
		{
			Name: "unblocking",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"blocking": false
}`,
			ExpectedMessage: makeContactWithActions([]msgcontents.ContactAction{
				msgcontents.ContactActionUnblock,
			}),
		},
		{
			Name: "following_and_unblocking",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"following": true,
	"blocking": false
}`,
			ExpectedMessage: makeContactWithActions([]msgcontents.ContactAction{
				msgcontents.ContactActionFollow,
				msgcontents.ContactActionUnblock,
			}),
		},
		{
			Name: "unfollowing_and_blocking",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"following": false,
	"blocking": true
}`,
			ExpectedMessage: makeContactWithActions([]msgcontents.ContactAction{
				msgcontents.ContactActionUnfollow,
				msgcontents.ContactActionBlock,
			}),
		},
		{
			Name: "following_and_blocking",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"following": true,
	"blocking": true
}`,
			ExpectedMessage: nil,
		},
		{
			Name: "unfollowing_and_unblocking",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"following": false,
	"blocking": false
}`,
			ExpectedMessage: makeContactWithActions([]msgcontents.ContactAction{
				msgcontents.ContactActionUnfollow,
				msgcontents.ContactActionUnblock,
			}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			marshaler := newMarshaler(t)

			msg, err := marshaler.Unmarshal(message.MustNewRawMessageContent([]byte(testCase.Content)))
			require.NoError(t, err)
			if testCase.ExpectedMessage != nil {
				require.Equal(t, testCase.ExpectedMessage, msg)
			} else {
				require.Equal(t, msgcontents.MustNewUnknown([]byte(testCase.Content)), msg)
			}
		})
	}
}

func TestMappingContactMarshal(t *testing.T) {
	iden := refs.MustNewIdentity("@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519")

	testCases := []struct {
		Name            string
		Actions         []msgcontents.ContactAction
		ExpectedContent string
	}{
		{
			Name: "following",
			Actions: []msgcontents.ContactAction{
				msgcontents.ContactActionFollow,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","following":true}`,
		},
		{
			Name: "unfollowing",
			Actions: []msgcontents.ContactAction{
				msgcontents.ContactActionUnfollow,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","following":false}`,
		},
		{
			Name: "blocking",
			Actions: []msgcontents.ContactAction{
				msgcontents.ContactActionBlock,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","blocking":true}`,
		},
		{
			Name: "unblocking",
			Actions: []msgcontents.ContactAction{
				msgcontents.ContactActionUnblock,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","blocking":false}`,
		},
		{
			Name: "unfollowing_and_blocking",
			Actions: []msgcontents.ContactAction{
				msgcontents.ContactActionUnfollow,
				msgcontents.ContactActionBlock,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","following":false,"blocking":true}`,
		},
		{
			Name: "following_and_unblocking",
			Actions: []msgcontents.ContactAction{
				msgcontents.ContactActionFollow,
				msgcontents.ContactActionUnblock,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","following":true,"blocking":false}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			msg := msgcontents.MustNewContact(iden, msgcontents.MustNewContactActions(testCase.Actions))

			marshaler := newMarshaler(t)

			raw, err := marshaler.Marshal(msg)
			require.NoError(t, err)

			require.Equal(
				t,
				testCase.ExpectedContent,
				string(raw.Bytes()),
			)
		})
	}
}
