package transport_test

import (
	"testing"

	msgcontents "github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestMappingContactUnmarshal(t *testing.T) {
	makeContactWithActions := func(actions []known.ContactAction) known.Contact {
		return known.MustNewContact(
			refs.MustNewIdentity("@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519"),
			known.MustNewContactActions(actions),
		)
	}

	testCases := []struct {
		Name            string
		Content         string
		ExpectedMessage known.KnownMessageContent
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
			ExpectedMessage: makeContactWithActions([]known.ContactAction{
				known.ContactActionFollow,
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
			ExpectedMessage: makeContactWithActions([]known.ContactAction{
				known.ContactActionUnfollow,
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
			ExpectedMessage: makeContactWithActions([]known.ContactAction{
				known.ContactActionBlock,
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
			ExpectedMessage: makeContactWithActions([]known.ContactAction{
				known.ContactActionUnblock,
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
			ExpectedMessage: makeContactWithActions([]known.ContactAction{
				known.ContactActionFollow,
				known.ContactActionUnblock,
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
			ExpectedMessage: makeContactWithActions([]known.ContactAction{
				known.ContactActionUnfollow,
				known.ContactActionBlock,
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
			ExpectedMessage: makeContactWithActions([]known.ContactAction{
				known.ContactActionUnfollow,
				known.ContactActionUnblock,
			}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			marshaler := newMarshaler(t)

			msg, err := marshaler.Unmarshal(message.MustNewRawMessageContent([]byte(testCase.Content)))
			if testCase.ExpectedMessage != nil {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedMessage, msg)
			} else {
				require.ErrorIs(t, err, msgcontents.ErrUnknownContent)
			}
		})
	}
}

func TestMappingContactMarshal(t *testing.T) {
	iden := refs.MustNewIdentity("@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519")

	testCases := []struct {
		Name            string
		Actions         []known.ContactAction
		ExpectedContent string
	}{
		{
			Name: "following",
			Actions: []known.ContactAction{
				known.ContactActionFollow,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","following":true}`,
		},
		{
			Name: "unfollowing",
			Actions: []known.ContactAction{
				known.ContactActionUnfollow,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","following":false}`,
		},
		{
			Name: "blocking",
			Actions: []known.ContactAction{
				known.ContactActionBlock,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","blocking":true}`,
		},
		{
			Name: "unblocking",
			Actions: []known.ContactAction{
				known.ContactActionUnblock,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","blocking":false}`,
		},
		{
			Name: "unfollowing_and_blocking",
			Actions: []known.ContactAction{
				known.ContactActionUnfollow,
				known.ContactActionBlock,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","following":false,"blocking":true}`,
		},
		{
			Name: "following_and_unblocking",
			Actions: []known.ContactAction{
				known.ContactActionFollow,
				known.ContactActionUnblock,
			},
			ExpectedContent: `{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","following":true,"blocking":false}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			msg := known.MustNewContact(iden, known.MustNewContactActions(testCase.Actions))

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
