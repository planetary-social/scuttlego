package formats

import (
	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content/transport"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// todo abstract away dependencies
func TestMarshaler(t *testing.T) {
	marshaler, err := transport.NewMarshaler(transport.DefaultMappings(), logging.NewLogrusLogger(logrus.New(), "test"))
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
