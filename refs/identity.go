package refs

import (
	"bytes"
	"encoding/base64"
	"strings"

	"github.com/boreq/errors"
	ssbidentity "github.com/planetary-social/go-ssb/identity"
)

const (
	identityPrefix = "@"
	identitySuffix = ".ed25519"

	maxIdentityLength = 100
)

type identity struct {
	s string
	k ssbidentity.Public
}

func newIdentityFromString(s string) (identity, error) {
	if !strings.HasPrefix(s, identityPrefix) {
		return identity{}, errors.New("invalid prefix")
	}

	if !strings.HasSuffix(s, identitySuffix) {
		return identity{}, errors.New("invalid suffix")
	}

	trimmed := s[len(identityPrefix) : len(s)-len(identitySuffix)]

	if base64.StdEncoding.DecodedLen(len(trimmed)) > maxIdentityLength {
		return identity{}, errors.New("encoded data is too long")
	}

	k, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return identity{}, errors.Wrap(err, "invalid base64")
	}

	public, err := ssbidentity.NewPublicFromBytes(k)
	if err != nil {
		return identity{}, errors.Wrap(err, "could not create a public identity")
	}

	return identity{s: s, k: public}, nil
}

func newIdentityFromPublic(public ssbidentity.Public) (identity, error) {
	if public.IsZero() {
		return identity{}, errors.New("zero value of identity")
	}

	encodedKey := base64.StdEncoding.EncodeToString(public.PublicKey())
	s := identityPrefix + encodedKey + identitySuffix
	return identity{s: s, k: public}, nil
}

func (i identity) Identity() ssbidentity.Public {
	return i.k
}

func (i identity) String() string {
	return i.s
}

func (i identity) Equal(o identity) bool {
	return bytes.Equal(i.k.PublicKey(), o.k.PublicKey()) // todo improve?
}

func (i identity) IsZero() bool {
	return i.k.IsZero()
}

type Identity struct {
	identity
}

func NewIdentity(s string) (Identity, error) {
	identity, err := newIdentityFromString(s)
	return Identity{identity}, err
}

func NewIdentityFromPublic(public ssbidentity.Public) (Identity, error) {
	identity, err := newIdentityFromPublic(public)
	return Identity{identity}, err
}

func MustNewIdentity(s string) Identity {
	identity, err := NewIdentity(s)
	if err != nil {
		panic(err)
	}
	return identity
}

func (i Identity) MainFeed() Feed {
	return Feed(i)
}

func (i Identity) Equal(o Identity) bool {
	return i.identity.Equal(o.identity)
}

type Feed struct {
	identity
}

func NewFeed(s string) (Feed, error) {
	identity, err := newIdentityFromString(s)
	return Feed{identity}, err
}

func MustNewFeed(s string) Feed {
	f, err := NewFeed(s)
	if err != nil {
		panic(err)
	}
	return f
}

func (f Feed) Equal(o Feed) bool {
	return f.identity.Equal(o.identity)
}
