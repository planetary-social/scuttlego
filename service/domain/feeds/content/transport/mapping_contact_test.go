package transport_test

import (
	"testing"

	"github.com/boreq/errors"
	msgcontents "github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestMappingContactUnmarshal(t *testing.T) {
	testCases := []struct {
		Name            string
		Content         string
		ExpectedActions []msgcontents.ContactAction
		ExpectedError   error
	}{
		{
			Name: "missing_action",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519"
}`,
			ExpectedError: errors.New("mapping 'contact' returned an error: could not unmarshal contact action: actions can not be empty"),
		},
		{
			Name: "following",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"following": true
}`,
			ExpectedActions: []msgcontents.ContactAction{
				msgcontents.ContactActionFollow,
			},
		},
		{
			Name: "unfollowing",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"following": false
}`,
			ExpectedActions: []msgcontents.ContactAction{
				msgcontents.ContactActionUnfollow,
			},
		},
		{
			Name: "blocking",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"blocking": true
}`,
			ExpectedActions: []msgcontents.ContactAction{
				msgcontents.ContactActionBlock,
			},
		},
		{
			Name: "unblocking",
			Content: `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"blocking": false
}`,
			ExpectedActions: []msgcontents.ContactAction{
				msgcontents.ContactActionUnblock,
			},
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
			ExpectedActions: []msgcontents.ContactAction{
				msgcontents.ContactActionFollow,
				msgcontents.ContactActionUnblock,
			},
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
			ExpectedActions: []msgcontents.ContactAction{
				msgcontents.ContactActionUnfollow,
				msgcontents.ContactActionBlock,
			},
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
			ExpectedError: errors.New("mapping 'contact' returned an error: could not unmarshal contact action: both follow and block are present"),
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
			ExpectedActions: []msgcontents.ContactAction{
				msgcontents.ContactActionUnfollow,
				msgcontents.ContactActionUnblock,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			marshaler := newMarshaler(t)

			msg, err := marshaler.Unmarshal(message.MustNewRawMessageContent([]byte(testCase.Content)))
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(
					t,
					msgcontents.MustNewContact(
						refs.MustNewIdentity("@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519"),
						msgcontents.MustNewContactActions(testCase.ExpectedActions),
					),
					msg,
				)
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
