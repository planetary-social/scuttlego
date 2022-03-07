package invites

import (
	"encoding/base64"
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/network"
	"github.com/planetary-social/go-ssb/refs"
	"strings"
)

type Invite struct {
	remote        refs.Identity
	address       network.Address
	secretKeySeed []byte
}

const (
	identitySeparator = ":"
	seedSeparator     = "~"
)

func NewInviteFromString(s string) (Invite, error) {
	seedString, err := readAfter(&s, seedSeparator)
	if err != nil {
		return Invite{}, errors.Wrap(err, "could not read the seed")
	}

	remoteString, err := readAfter(&s, identitySeparator)
	if err != nil {
		return Invite{}, errors.Wrap(err, "could not read the remote identity")
	}

	addressString := s

	seed, err := base64.StdEncoding.DecodeString(seedString)
	if err != nil {
		return Invite{}, errors.Wrap(err, "could not decode the seed")
	}

	remote, err := refs.NewIdentity(remoteString)
	if err != nil {
		return Invite{}, errors.Wrap(err, "invalid identity")
	}

	address := network.NewAddress(addressString)

	return Invite{
		remote:        remote,
		address:       address,
		secretKeySeed: seed,
	}, nil
}

func readAfter(s *string, separator string) (string, error) {
	index := strings.LastIndex(*s, separator)
	if index < 0 {
		return "", errors.New("could not find the separator")
	}
	result := (*s)[index+1:]
	*s = (*s)[:index]
	return result, nil
}

func MustNewInviteFromString(s string) Invite {
	i, err := NewInviteFromString(s)
	if err != nil {
		panic(err)
	}
	return i
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
