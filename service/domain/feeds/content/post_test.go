package content_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestPostImplementsBlobReferencer(t *testing.T) {
	post, err := content.NewPost(nil)
	require.NoError(t, err)
	require.Implements(t, new(blobs.BlobReferencer), post)
}

func TestPostBlobs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		about, err := content.NewPost(nil)
		require.NoError(t, err)
		require.Empty(t, about.Blobs())
	})

	t.Run("not_empty", func(t *testing.T) {
		blobRef := fixtures.SomeRefBlob()
		about, err := content.NewPost([]refs.Blob{blobRef})
		require.NoError(t, err)
		require.Equal(t, []refs.Blob{blobRef}, about.Blobs())
	})
}
