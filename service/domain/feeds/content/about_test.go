package content_test

import (
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestAboutImplementsBlobReferencer(t *testing.T) {
	about, err := content.NewAbout(nil)
	require.NoError(t, err)
	require.Implements(t, new(blobs.BlobReferencer), about)
}

func TestAboutBlobs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		about, err := content.NewAbout(nil)
		require.NoError(t, err)
		require.Empty(t, about.Blobs())
	})

	t.Run("not_empty", func(t *testing.T) {
		blobRef := fixtures.SomeRefBlob()
		about, err := content.NewAbout(&blobRef)
		require.NoError(t, err)
		require.Equal(t, []refs.Blob{blobRef}, about.Blobs())
	})
}
