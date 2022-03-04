package invites

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/network"
	"github.com/planetary-social/go-ssb/refs"
	"go.cryptoscope.co/ssb/invite"
)

type Invite struct {
	remote        refs.Identity
	address       network.Address
	secretKeySeed []byte
}

func NewInviteFromString(s string) (Invite, error) {
	token, err := invite.ParseLegacyToken(s)
	if err != nil {
		return Invite{}, errors.Wrap(err, "invalid invite string")
	}

	remote, err := refs.NewIdentity(token.Peer.Ref())
	if err != nil {
		return Invite{}, errors.Wrap(err, "invalid identity")
	}

	return Invite{
		remote:        remote,
		address:       network.NewAddress(token.Address.String()),
		secretKeySeed: token.Seed[:],
	}, nil
}

func (i Invite) Remote() refs.Identity {
	return i.remote
}

func (i Invite) Address() network.Address {
	return i.address
}

func (i Invite) SecretKeySeed() []byte {
	return i.secretKeySeed
}
