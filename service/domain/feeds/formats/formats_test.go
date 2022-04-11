package formats

import (
	"strings"
	"testing"

	"github.com/planetary-social/go-ssb/service/domain/transport/boxstream"
	"github.com/stretchr/testify/require"
)

func TestMessageHMAC_BytesReturnsNilForDefaultHMAC(t *testing.T) {
	hmac := NewDefaultMessageHMAC()
	require.True(t, hmac.Bytes() == nil)
}

func TestMessageHMAC_ConstructorCanCreateDefaultHMAC(t *testing.T) {
	hmac, err := NewMessageHMAC(nil)
	require.NoError(t, err)

	require.True(t, hmac.Bytes() == nil)
}

func TestMessageHMAC_Bytes(t *testing.T) {
	sliceOf32Bytes := []byte(strings.Repeat("a", 32))

	key, err := boxstream.NewNetworkKey(sliceOf32Bytes)
	require.NoError(t, err)

	require.Len(t, key.Bytes(), 32)
}
