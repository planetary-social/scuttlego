package formats

import (
	"testing"
	"time"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content/transport"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
	"github.com/stretchr/testify/require"
)

// todo abstract away dependencies
func TestMarshaler(t *testing.T) {
	logger := fixtures.SomeLogger()
	marshaler, err := transport.NewMarshaler(transport.DefaultMappings(), logger)
	require.NoError(t, err)

	f := NewScuttlebutt(marshaler)

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	unsignedMessage, err := message.NewUnsignedMessage(
		nil,
		message.FirstSequence,
		authorRef,
		authorRef.MainFeed(),
		time.Now(),
		content.MustNewContact(fixtures.SomeRefAuthor(), content.ContactActionFollow),
	)
	require.NoError(t, err)

	_, err = f.Sign(unsignedMessage, author)
	require.NoError(t, err)
}

// todo abstract away dependencies
func TestMarshalerPrevious(t *testing.T) {
	logger := fixtures.SomeLogger()
	marshaler, err := transport.NewMarshaler(transport.DefaultMappings(), logger)
	require.NoError(t, err)

	f := NewScuttlebutt(marshaler)

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
		content.MustNewContact(fixtures.SomeRefAuthor(), content.ContactActionFollow),
	)
	require.NoError(t, err)

	_, err = f.Sign(unsignedMessage, author)
	require.NoError(t, err)
}
