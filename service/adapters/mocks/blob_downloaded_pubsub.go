package mocks

import (
	"context"

	"github.com/planetary-social/scuttlego/service/app/queries"
)

type BlobDownloadedPubSubMock struct {
}

func NewBlobDownloadedPubSubMock() *BlobDownloadedPubSubMock {
	return &BlobDownloadedPubSubMock{}
}

func (b BlobDownloadedPubSubMock) Subscribe(ctx context.Context) <-chan queries.BlobDownloaded {
	ch := make(chan queries.BlobDownloaded)
	close(ch)
	return ch
}
