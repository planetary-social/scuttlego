package boxstream_test

import (
	"strings"
	"testing"

	"github.com/planetary-social/go-ssb/service/domain/transport/boxstream"
	"github.com/stretchr/testify/require"
)

func TestNetworkKey_Bytes(t *testing.T) {
	sliceOf32Bytes := []byte(strings.Repeat("a", 32))

	key, err := boxstream.NewNetworkKey(sliceOf32Bytes)
	require.NoError(t, err)

	require.Len(t, key.Bytes(), 32)
}
