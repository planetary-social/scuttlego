package pubsub

import (
	"context"

	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type BlobDownloadedPubSub struct {
	pubsub *GoChannelPubSub[queries.BlobDownloaded]
}

func NewBlobDownloadedPubSub() *BlobDownloadedPubSub {
	return &BlobDownloadedPubSub{
		pubsub: NewGoChannelPubSub[queries.BlobDownloaded](),
	}
}

func (m *BlobDownloadedPubSub) Publish(blob refs.Blob, size blobs.Size) {
	m.pubsub.Publish(
		queries.BlobDownloaded{
			Id:   blob,
			Size: size,
		},
	)
}

func (m *BlobDownloadedPubSub) Subscribe(ctx context.Context) <-chan queries.BlobDownloaded {
	return m.pubsub.Subscribe(ctx)
}
