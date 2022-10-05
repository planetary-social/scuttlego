package aliases

import (
	"crypto/ed25519"
	"fmt"
	"strings"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type RegistrationMessage struct {
	alias Alias
	user  refs.Identity
	room  refs.Identity
}

func NewRegistrationMessage(alias Alias, user refs.Identity, room refs.Identity) (RegistrationMessage, error) {
	if alias.IsZero() {
		return RegistrationMessage{}, errors.New("zero value of alias")
	}

	if user.IsZero() {
		return RegistrationMessage{}, errors.New("zero value of user")
	}

	if room.IsZero() {
		return RegistrationMessage{}, errors.New("zero value of room")
	}

	return RegistrationMessage{
		alias: alias,
		user:  user,
		room:  room,
	}, nil
}

func (r RegistrationMessage) String() string {
	var message strings.Builder
	message.WriteString("=room-alias-registration:")
	message.WriteString(r.room.String())
	message.WriteString(":")
	message.WriteString(r.user.String())
	message.WriteString(":")
	message.WriteString(r.alias.String())
	return message.String()
}

func (r RegistrationMessage) IsZero() bool {
	return r.alias.IsZero()
}

type RegistrationSignature struct {
	signature []byte
}

func NewRegistrationSignature(msg RegistrationMessage, private identity.Private) (RegistrationSignature, error) {
	if msg.IsZero() {
		return RegistrationSignature{}, errors.New("zero value of registration message")
	}

	if private.IsZero() {
		return RegistrationSignature{}, errors.New("zero value of identity")
	}

	public, err := refs.NewIdentityFromPublic(private.Public())
	if err != nil {
		return RegistrationSignature{}, errors.New("failed to create a public identity")
	}

	if !public.Equal(msg.user) {
		return RegistrationSignature{}, errors.New("private identity doesn't match user identity from the message")
	}

	signature := ed25519.Sign(private.PrivateKey(), []byte(msg.String()))
	return RegistrationSignature{
		signature: signature,
	}, nil
}

func (s RegistrationSignature) Bytes() []byte {
	tmp := make([]byte, len(s.signature))
	copy(tmp, s.signature)
	return tmp
}

func (s RegistrationSignature) IsZero() bool {
	return len(s.signature) == 0
}

type Alias struct {
	s string
}

func NewAlias(s string) (Alias, error) {
	if s == "" {
		return Alias{}, errors.New("empty string")
	}

	if len(s) > 63 {
		return Alias{}, errors.New("string too long")
	}

	for _, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}

		if r >= 'a' && r <= 'z' {
			continue
		}

		return Alias{}, fmt.Errorf("invalid character: '%c'", r)
	}

	return Alias{s: s}, nil
}

func MustNewAlias(s string) Alias {
	v, err := NewAlias(s)
	if err != nil {
		panic(err)
	}
	return v
}

func (a Alias) IsZero() bool {
	return a == Alias{}
}

func (a Alias) String() string {
	return a.s
}

// AliasEndpointURL should be a valid URL which can be used to get information
// about this alias using HTTP requests. The URL can be similar to either:
// - https://somealias.example.com
// - https://example.com/alias/somealias
type AliasEndpointURL struct {
	s string
}

func NewAliasEndpointURL(s string) (AliasEndpointURL, error) {
	if s == "" {
		return AliasEndpointURL{}, errors.New("empty string")
	}
	return AliasEndpointURL{s: s}, nil
}

func (a AliasEndpointURL) String() string {
	return a.s
}
