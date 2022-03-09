package formats

import (
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	transport2 "github.com/planetary-social/go-ssb/service/domain/feeds/content/transport"
	message2 "github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"testing"
	"time"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/stretchr/testify/require"
)

// todo abstract away dependencies
func TestMarshaler(t *testing.T) {
	logger := fixtures.SomeLogger()
	marshaler, err := transport2.NewMarshaler(transport2.DefaultMappings(), logger)
	require.NoError(t, err)

	f := NewScuttlebutt(marshaler)

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	unsignedMessage, err := message2.NewUnsignedMessage(
		nil,
		message2.FirstSequence,
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
	marshaler, err := transport2.NewMarshaler(transport2.DefaultMappings(), logger)
	require.NoError(t, err)

	f := NewScuttlebutt(marshaler)

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	previous := fixtures.SomeRefMessage()

	unsignedMessage, err := message2.NewUnsignedMessage(
		&previous,
		message2.MustNewSequence(2),
		authorRef,
		authorRef.MainFeed(),
		time.Now(),
		content.MustNewContact(fixtures.SomeRefAuthor(), content.ContactActionFollow),
	)
	require.NoError(t, err)

	_, err = f.Sign(unsignedMessage, author)
	require.NoError(t, err)
}
