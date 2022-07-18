package identity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublic_Equal(t *testing.T) {
	a, err := NewPrivate()
	require.NoError(t, err)

	b, err := NewPrivate()
	require.NoError(t, err)

	require.True(t, a.Public().Equal(a.Public()))
	require.False(t, a.Public().Equal(b.Public()))
}
