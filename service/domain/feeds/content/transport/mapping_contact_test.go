package transport_test

import (
	"testing"

	msgcontents "github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestMappingContactUnmarshal(t *testing.T) {
	marshaler := newMarshaler(t)

	content := `
{
	"type": "contact",
	"contact": "@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519",
	"following": true
}`

	msg, err := marshaler.Unmarshal(message.MustNewRawMessageContent([]byte(content)))
	require.NoError(t, err)

	require.Equal(
		t,
		msgcontents.MustNewContact(
			refs.MustNewIdentity("@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519"),
			msgcontents.ContactActionFollow,
		),
		msg,
	)
}

func TestMappingContactMarshal(t *testing.T) {
	marshaler := newMarshaler(t)

	msg := msgcontents.MustNewContact(
		refs.MustNewIdentity("@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519"),
		msgcontents.ContactActionFollow,
	)

	raw, err := marshaler.Marshal(msg)
	require.NoError(t, err)

	require.Equal(
		t,
		`{"type":"contact","contact":"@sxlUkN7dW/qZ23Wid6J1IAnqWEJ3V13dT6TaFtn5LTc=.ed25519","following":true,"blocking":false}`,
		string(raw.Bytes()),
	)
}
