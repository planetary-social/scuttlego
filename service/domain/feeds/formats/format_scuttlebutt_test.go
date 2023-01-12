package formats

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/transport"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestScuttlebutt_MessageCanBeSignedAndThenVerified(t *testing.T) {
	f := newScuttlebuttFormat(t, NewDefaultMessageHMAC())

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	unsignedMessage, err := message.NewUnsignedMessage(
		nil,
		message.NewFirstSequence(),
		authorRef,
		authorRef.MainFeed(),
		time.Now(),
		someContent(),
	)
	require.NoError(t, err)

	msgFromSign, err := f.Sign(unsignedMessage, author)
	require.NoError(t, err)

	msgFromVerify, err := f.Verify(msgFromSign.Raw())
	require.NoError(t, err)

	require.Equal(t, msgFromSign, msgFromVerify)
}

func TestScuttlebutt_MessageCanBeSignedAndThenVerifiedIfItReferencesAPreviousMessage(t *testing.T) {
	f := newScuttlebuttFormat(t, NewDefaultMessageHMAC())

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	previous := fixtures.SomeRefMessage()

	unsignedMessage, err := message.NewUnsignedMessage(
		&previous,
		message.MustNewSequence(2),
		authorRef,
		authorRef.MainFeed(),
		time.Now(),
		someContent(),
	)
	require.NoError(t, err)

	msgFromSign, err := f.Sign(unsignedMessage, author)
	require.NoError(t, err)

	msgFromVerify, err := f.Verify(msgFromSign.Raw())
	require.NoError(t, err)

	require.Equal(t, msgFromSign, msgFromVerify)
}

func TestScuttlebutt_SettingHMACMakesMessagesIncompatibile(t *testing.T) {
	hmac, err := NewMessageHMAC([]byte("somehmacthatislongenoughblablabl"))
	require.NoError(t, err)

	f := newScuttlebuttFormat(t, hmac)

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	unsignedMessage, err := message.NewUnsignedMessage(
		nil,
		message.NewFirstSequence(),
		authorRef,
		authorRef.MainFeed(),
		time.Now(),
		someContent(),
	)
	require.NoError(t, err)

	msg, err := f.Sign(unsignedMessage, author)
	require.NoError(t, err)

	_, err = f.Verify(msg.Raw())
	require.NoError(t, err)

	defaultFormat := newScuttlebuttFormat(t, NewDefaultMessageHMAC())
	_, err = defaultFormat.Verify(msg.Raw())
	require.Contains(t, err.Error(), "invalid signature")
}

func TestScuttlebutt_LoadAndVerifyReturnIdenticalResults(t *testing.T) {
	f := newScuttlebuttFormat(t, NewDefaultMessageHMAC())

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	previous := fixtures.SomeRefMessage()

	unsignedMessage, err := message.NewUnsignedMessage(
		&previous,
		message.MustNewSequence(2),
		authorRef,
		authorRef.MainFeed(),
		time.Now(),
		someContent(),
	)
	require.NoError(t, err)

	msg, err := f.Sign(unsignedMessage, author)
	require.NoError(t, err)

	verifiedRawMessage, err := message.NewVerifiedRawMessage(msg.Raw().Bytes())
	require.NoError(t, err)

	msgFromLoadWithoutId, err := f.Load(verifiedRawMessage)
	require.NoError(t, err)

	msgFromLoad, err := message.NewMessageFromMessageWithoutId(msg.Id(), msgFromLoadWithoutId)
	require.NoError(t, err)

	msgFromVerify, err := f.Verify(msg.Raw())
	require.NoError(t, err)

	require.Equal(t, msg, msgFromVerify)
	require.Equal(t, msgFromLoad, msgFromVerify)
}

func someContent() message.RawMessageContent {
	return message.MustNewRawMessageContent([]byte(`{"type": "something"}`))
}

func newScuttlebuttFormat(t *testing.T, hmac MessageHMAC) *Scuttlebutt {
	logger := fixtures.SomeLogger()
	marshaler, err := transport.NewMarshaler(transport.DefaultMappings(), logger)
	require.NoError(t, err)

	return NewScuttlebutt(marshaler, hmac)
}
