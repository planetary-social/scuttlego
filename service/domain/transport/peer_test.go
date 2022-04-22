package transport_test

import (
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport"
	"github.com/stretchr/testify/require"
)

func TestPeer_String(t *testing.T) {
	peer := transport.NewPeer(fixtures.SomePublicIdentity(), nil)

	identityRef, err := refs.NewIdentityFromPublic(peer.Identity())
	require.NoError(t, err)

	require.Equal(t, "<peer "+identityRef.String()+">", peer.String())
}
