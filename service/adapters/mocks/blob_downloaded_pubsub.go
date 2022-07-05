package mocks

import (
	"context"

	"github.com/planetary-social/scuttlego/service/app/queries"
)

type BlobDownloadedPubSubMock struct {
	CallsCount int
}

func NewBlobDownloadedPubSubMock() *BlobDownloadedPubSubMock {
	return &BlobDownloadedPubSubMock{}
}

func (b BlobDownloadedPubSubMock) Subscribe(ctx context.Context) <-chan queries.BlobDownloaded {
	b.CallsCount++
	ch := make(chan queries.BlobDownloaded)
	close(ch)
	return ch
}
