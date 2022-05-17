package bolt_test

import (
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/adapters/bolt"
	"github.com/planetary-social/go-ssb/service/domain/feeds"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestBlobRepository_Put(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		repository := bolt.NewBlobRepository(tx)
		blobsToSave := feeds.NewBlobsToSave(
			fixtures.SomeRefFeed(),
			fixtures.SomeRefMessage(),
			[]refs.Blob{
				fixtures.SomeRefBlob(),
			},
		)
		return repository.Put(blobsToSave)
	})
	require.NoError(t, err)
}
