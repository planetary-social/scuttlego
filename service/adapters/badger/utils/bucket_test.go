package utils_test

import (
	"fmt"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/stretchr/testify/require"
)

func BenchmarkBucket_KeyInBucket(b *testing.B) {
	for _, numberOfComponents := range []int{1, 3, 5, 7, 9} {
		b.Run(fmt.Sprintf("components-%d", numberOfComponents), func(b *testing.B) {
			var components []utils.KeyComponent
			for i := 0; i < numberOfComponents; i++ {
				components = append(components, utils.MustNewKeyComponent(fixtures.SomeBytes()))
			}

			prefix := utils.MustNewKey(components...)

			bucket, err := utils.NewBucket(&badger.Txn{}, prefix)
			require.NoError(b, err)

			item := newItemMock(prefix.Append(utils.MustNewKeyComponent(fixtures.SomeBytes())).Bytes())

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := bucket.KeyInBucket(item)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

type itemMock struct {
	key []byte
}

func newItemMock(key []byte) *itemMock {
	return &itemMock{key: key}
}

func (i itemMock) KeyCopy(bytes []byte) []byte {
	return i.key
}

func (i itemMock) ValueCopy(bytes []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (i itemMock) DangerousKey(f func(key []byte) error) error {
	//TODO implement me
	panic("implement me")
}

func (i itemMock) DangerousValue(f func(value []byte) error) error {
	//TODO implement me
	panic("implement me")
}
