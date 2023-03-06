package replication_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestStorageBlobsThatShouldBePushedProvider_GetBlobsThatShouldBePushed(t *testing.T) {
	blob1 := fixtures.SomeRefBlob()
	blob2 := fixtures.SomeRefBlob()

	now := time.Now()

	newMessage := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		nil,
		message.NewFirstSequence(),
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefFeed(),
		now,
		fixtures.SomeContent(),
		fixtures.SomeRawMessage(),
	)
	oldMessage := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		nil,
		message.NewFirstSequence(),
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefFeed(),
		now.Add(-24*time.Hour),
		fixtures.SomeContent(),
		fixtures.SomeRawMessage(),
	)

	testCases := []struct {
		Name        string
		Blobs       []replication.MessageBlobs
		MustContain []refs.Blob
	}{
		{
			Name: "must_contain_new_message",
			Blobs: []replication.MessageBlobs{
				{
					Message: newMessage,
					Blobs: []refs.Blob{
						blob1,
					},
				},
				{
					Message: oldMessage,
					Blobs: []refs.Blob{
						blob2,
					},
				},
			},
			MustContain: []refs.Blob{
				blob1,
			},
		},
		{
			Name: "duplicates_are_deduplicated",
			Blobs: []replication.MessageBlobs{
				{
					Message: newMessage,
					Blobs: []refs.Blob{
						blob1,
						blob2,
					},
				},
				{
					Message: newMessage,
					Blobs: []refs.Blob{
						blob1,
						blob2,
					},
				},
			},
			MustContain: []refs.Blob{
				blob1,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			blobsRepository := newBlobsRepositoryMock()
			identity := fixtures.SomePublicIdentity()
			currentTimeProvider := mocks.NewCurrentTimeProviderMock()

			blobsRepository.GetFeedBlobsReturnValue = testCase.Blobs
			currentTimeProvider.CurrentTime = now

			p, err := replication.NewStorageBlobsThatShouldBePushedProvider(blobsRepository, identity, currentTimeProvider)
			require.NoError(t, err)

			got, err := p.GetBlobsThatShouldBePushed()
			require.NoError(t, err)

			for _, v := range testCase.MustContain {
				require.Contains(t, got, v)
			}
		})
	}
}

func TestCacheBlobsThatShouldBePushedProvider_GetBlobsThatShouldBePushedCachesTheValueAndCallsTheUnderlyingProviderOnce(t *testing.T) {
	provider := newProviderMock()
	cacheProvider := replication.NewCacheBlobsThatShouldBePushedProvider(provider)

	_, err := cacheProvider.GetBlobsThatShouldBePushed()
	require.NoError(t, err)

	_, err = cacheProvider.GetBlobsThatShouldBePushed()
	require.NoError(t, err)

	require.Equal(t, 1, provider.GetBlobsThatShouldBePushedCalls)
}

type providerMock struct {
	GetBlobsThatShouldBePushedCalls int
}

func newProviderMock() *providerMock {
	return &providerMock{}
}

func (p *providerMock) GetBlobsThatShouldBePushed() ([]refs.Blob, error) {
	p.GetBlobsThatShouldBePushedCalls++
	return nil, nil
}

type blobsRepositoryMock struct {
	GetFeedBlobsReturnValue []replication.MessageBlobs
}

func newBlobsRepositoryMock() *blobsRepositoryMock {
	return &blobsRepositoryMock{}
}

func (b blobsRepositoryMock) GetFeedBlobs(id refs.Feed) ([]replication.MessageBlobs, error) {
	return b.GetFeedBlobsReturnValue, nil
}
