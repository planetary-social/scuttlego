package identity

import (
	"crypto/ed25519"
	"crypto/rand"

	"github.com/boreq/errors"
)

type Public struct {
	key ed25519.PublicKey
}

func NewPublicFromPrivate(private Private) Public {
	return Public{private.key.Public().(ed25519.PublicKey)}
}

func NewPublicFromBytes(b []byte) (Public, error) {
	if len(b) != ed25519.PublicKeySize {
		return Public{}, errors.New("invalid public key length")
	}

	return Public{b}, nil
}

func (p Public) PublicKey() ed25519.PublicKey {
	return p.key
}

func (p Public) IsZero() bool {
	return len(p.key) == 0
}

type Private struct {
	key ed25519.PrivateKey
}

func NewPrivate() (Private, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return Private{}, errors.Wrap(err, "could not generate a secret key")
	}

	return Private{
		key: privateKey,
	}, nil
}

func NewPrivateFromSeed(seed []byte) (Private, error) {
	if len(seed) != ed25519.SeedSize {
		return Private{}, errors.New("invalid seed size")
	}

	return Private{ed25519.NewKeyFromSeed(seed)}, nil
}

func (p Private) Public() Public {
	return NewPublicFromPrivate(p)
}

func (p Private) PrivateKey() ed25519.PrivateKey {
	return p.key
}
