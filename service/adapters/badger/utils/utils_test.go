package utils_test

import (
	"fmt"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/stretchr/testify/require"
)

func TestNewKeyFromBytes_CorrectlyUnmarshalsKeys(t *testing.T) {
	a := utils.MustNewKeyComponent(fixtures.SomeBytes())
	b := utils.MustNewKeyComponent(fixtures.SomeBytes())
	c := utils.MustNewKeyComponent(fixtures.SomeBytes())

	key, err := utils.NewKey(a, b, c)
	require.NoError(t, err)

	marshaledKey := key.Bytes()

	unmarshaledKey, err := utils.NewKeyFromBytes(marshaledKey)
	require.NoError(t, err)

	require.Equal(t, key, unmarshaledKey)
	require.Equal(t,
		[]utils.KeyComponent{
			a,
			b,
			c,
		},
		unmarshaledKey.Components(),
	)
}

func TestNewKeyFromBytes_DetectsMalformedData(t *testing.T) {
	_, err := utils.NewKeyFromBytes([]byte("test"))
	require.EqualError(t, err, "remaining data length < next sequence length")
}

func BenchmarkNewKeyFromBytes(b *testing.B) {
	for _, numberOfComponents := range []int{1, 3, 5, 7, 9} {
		b.Run(fmt.Sprintf("components-%d", numberOfComponents), func(b *testing.B) {
			var components []utils.KeyComponent
			for i := 0; i < numberOfComponents; i++ {
				components = append(components, utils.MustNewKeyComponent(fixtures.SomeBytes()))
			}

			key, err := utils.NewKey(components...)
			require.NoError(b, err)

			marshaledKey := key.Bytes()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := utils.NewKeyFromBytes(marshaledKey)
				require.NoError(b, err)
			}
		})
	}
}
