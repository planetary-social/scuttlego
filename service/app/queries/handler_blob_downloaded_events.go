package queries

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type BlobDownloadedSubscriber interface {
	Subscribe(ctx context.Context) <-chan BlobDownloaded
}

type BlobDownloaded struct {
	Id   refs.Blob
	Size blobs.Size
}

type BlobDownloadedEventsHandler struct {
	subscriber BlobDownloadedSubscriber
}

func NewBlobDownloadedEventsHandler(subscriber BlobDownloadedSubscriber) *BlobDownloadedEventsHandler {
	return &BlobDownloadedEventsHandler{subscriber: subscriber}
}

func (h *BlobDownloadedEventsHandler) Handle(ctx context.Context) <-chan BlobDownloaded {
	return h.subscriber.Subscribe(ctx)
}
