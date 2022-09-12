package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	s := NewSet[int]()

	require.False(t, s.Contains(0))
	require.Equal(t, 0, s.Len())
	require.Len(t, s.List(), 0)

	s.Put(0)

	require.True(t, s.Contains(0))
	require.Equal(t, 1, s.Len())
	require.Equal(t, []int{0}, s.List())
}
