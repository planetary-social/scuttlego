package aliases_test

import (
	"strings"
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/rooms/aliases"
	"github.com/stretchr/testify/require"
)

func TestNewAlias(t *testing.T) {
	testCases := []struct {
		Alias         string
		ExpectedError error
	}{
		{
			Alias:         "validalias123",
			ExpectedError: nil,
		},
		{
			Alias:         "invalid-alias123",
			ExpectedError: errors.New("invalid character: '-'"),
		},
		{
			Alias:         "",
			ExpectedError: errors.New("empty string"),
		},
		{
			Alias:         strings.Repeat("a", 100),
			ExpectedError: errors.New("string too long"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Alias, func(t *testing.T) {
			alias, err := aliases.NewAlias(testCase.Alias)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				require.Equal(t, testCase.Alias, alias.String())
				require.False(t, alias.IsZero())
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}

func TestRegistrationMessage_ProducesExpectedMessageString(t *testing.T) {
	alias, err := aliases.NewAlias("somealias")
	require.NoError(t, err)

	user, err := refs.NewIdentity("@gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=.ed25519")
	require.NoError(t, err)

	room, err := refs.NewIdentity("@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519")
	require.NoError(t, err)

	message, err := aliases.NewRegistrationMessage(alias, user, room)
	require.NoError(t, err)

	require.Equal(t,
		"=room-alias-registration:@Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.ed25519:@gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=.ed25519:somealias",
		message.String(),
	)
}

func TestNewRegistrationSignature_PrivateIdentityMustMatchUserFromTheMessage(t *testing.T) {
	identity := fixtures.SomePrivateIdentity()

	message, err := aliases.NewRegistrationMessage(fixtures.SomeAlias(), fixtures.SomeRefIdentity(), fixtures.SomeRefIdentity())
	require.NoError(t, err)

	_, err = aliases.NewRegistrationSignature(message, identity)
	require.EqualError(t, err, "private identity doesn't match user identity from the message")
}

func TestRegistrationSignature_CreatesNonZeroSignatures(t *testing.T) {
	identity := fixtures.SomePrivateIdentity()

	message, err := aliases.NewRegistrationMessage(
		fixtures.SomeAlias(),
		refs.MustNewIdentityFromPublic(identity.Public()),
		fixtures.SomeRefIdentity(),
	)
	require.NoError(t, err)

	signature, err := aliases.NewRegistrationSignature(message, identity)
	require.NoError(t, err)
	require.NotEmpty(t, signature.Bytes())
	require.False(t, signature.IsZero())
}

func TestNewAliasEndpointURL(t *testing.T) {
	testCases := []struct {
		String        string
		ExpectedError error
	}{
		{
			String:        "https://somealias.example.com",
			ExpectedError: nil,
		},
		{
			String:        "https://example.com/alias/somealias",
			ExpectedError: nil,
		},
		{
			String:        "somealias", // invalid? too hard to validate tbh
			ExpectedError: nil,
		},
		{
			String:        "",
			ExpectedError: errors.New("empty string"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.String, func(t *testing.T) {
			url, err := aliases.NewAliasEndpointURL(testCase.String)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				require.Equal(t, testCase.String, url.String())
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}
