package feeds

import (
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type BlobsToSave struct {
	feed    refs.Feed
	message refs.Message
	blobs   []refs.Blob
}

func NewBlobsToSave(feed refs.Feed, message refs.Message, blobs []refs.Blob) BlobsToSave {
	return BlobsToSave{
		feed:    feed,
		message: message,
		blobs:   blobs,
	}
}

func (b BlobsToSave) Feed() refs.Feed {
	return b.feed
}

func (b BlobsToSave) Message() refs.Message {
	return b.message
}

func (b BlobsToSave) Blobs() []refs.Blob {
	return b.blobs
}
