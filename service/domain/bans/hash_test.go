package bans_test

import (
	"bytes"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/stretchr/testify/require"
)

func TestNewHash(t *testing.T) {
	_, err := bans.NewHash(nil)
	require.Error(t, err)

	_, err = bans.NewHash(fixtures.SomeBytesOfLength(0))
	require.Error(t, err)

	_, err = bans.NewHash(fixtures.SomeBytesOfLength(32))
	require.NoError(t, err)

	_, err = bans.NewHash(fixtures.SomeBytesOfLength(64))
	require.Error(t, err)
}

func TestHash_Bytes(t *testing.T) {
	b := fixtures.SomeBytesOfLength(32)

	h, err := bans.NewHash(b)
	require.NoError(t, err)

	require.True(t, bytes.Equal(h.Bytes(), b))
}

func TestHash_IsZero(t *testing.T) {
	h, err := bans.NewHash(fixtures.SomeBytesOfLength(32))
	require.NoError(t, err)

	require.False(t, h.IsZero())
	require.True(t, bans.Hash{}.IsZero())
}
