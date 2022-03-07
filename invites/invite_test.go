package invites_test

import (
	"testing"

	"github.com/planetary-social/go-ssb/invites"
	"github.com/stretchr/testify/require"
)

func TestInviteString(t *testing.T) {
	inviteString := "one.planetary.pub:8008:@CIlwTOK+m6v1hT2zUVOCJvvZq7KE/65ErN6yA2yrURY=.ed25519~KVvak/aZeQJQUrn1imLIvwU+EVTkCzGW8TJWTmK8lOk="

	invite, err := invites.NewInviteFromString(inviteString)
	require.NoError(t, err)

	require.Equal(t, "@CIlwTOK+m6v1hT2zUVOCJvvZq7KE/65ErN6yA2yrURY=.ed25519", invite.Remote().String())
	require.Equal(t, "one.planetary.pub:8008", invite.Address().String())
	require.NotEmpty(t, invite.SecretKeySeed())
}

func TestInviteString_empty(t *testing.T) {
	_, err := invites.NewInviteFromString("")
	require.Error(t, err)
}
