package replication

import (
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const (
	alwaysPushBlobsForMessagesYoungerThan = 6 * time.Hour
	pushBlobsForNumberOfRandomMessages    = 5
)

type MessageBlobs struct {
	Message message.Message
	Blobs   []refs.Blob
}

type BlobsRepository interface {
	GetFeedBlobs(id refs.Feed) ([]MessageBlobs, error)
}

type CurrentTimeProvider interface {
	Get() time.Time
}

type CacheBlobsThatShouldBePushedProvider struct {
	provider BlobsThatShouldBePushedProvider

	refreshCacheEvery time.Duration

	cache          []refs.Blob
	cacheTimestamp time.Time
	cacheLock      sync.Mutex
}

func NewCacheBlobsThatShouldBePushedProvider(provider BlobsThatShouldBePushedProvider) *CacheBlobsThatShouldBePushedProvider {
	return &CacheBlobsThatShouldBePushedProvider{
		provider: provider,

		refreshCacheEvery: 15 * time.Second,
	}
}

func (c *CacheBlobsThatShouldBePushedProvider) GetBlobsThatShouldBePushed() ([]refs.Blob, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if c.cacheTimestamp.IsZero() || time.Since(c.cacheTimestamp) > c.refreshCacheEvery {
		tmp, err := c.provider.GetBlobsThatShouldBePushed()
		if err != nil {
			return nil, errors.Wrap(err, "error refreshing the cache")
		}

		c.cache = tmp
		c.cacheTimestamp = time.Now()
	}

	return c.cache, nil
}

type StorageBlobsThatShouldBePushedProvider struct {
	blobs               BlobsRepository
	localRef            refs.Identity
	currentTimeProvider CurrentTimeProvider
}

func NewStorageBlobsThatShouldBePushedProvider(
	blobs BlobsRepository,
	local identity.Public,
	currentTimeProvider CurrentTimeProvider,
) (*StorageBlobsThatShouldBePushedProvider, error) {
	localRef, err := refs.NewIdentityFromPublic(local)
	if err != nil {
		return nil, errors.Wrap(err, "error creating local identity")
	}

	return &StorageBlobsThatShouldBePushedProvider{
		blobs:               blobs,
		localRef:            localRef,
		currentTimeProvider: currentTimeProvider,
	}, nil
}

func (s *StorageBlobsThatShouldBePushedProvider) GetBlobsThatShouldBePushed() ([]refs.Blob, error) {
	feedBlobs, err := s.blobs.GetFeedBlobs(s.localRef.MainFeed())
	if err != nil {
		return nil, errors.Wrap(err, "error getting blobs")
	}

	result := make(map[string]refs.Blob)

	for _, feedBlob := range feedBlobs {
		if feedBlob.Message.Timestamp().After(s.currentTimeProvider.Get().Add(-alwaysPushBlobsForMessagesYoungerThan)) {
			for _, blob := range feedBlob.Blobs {
				result[blob.String()] = blob
			}
		}
	}

	internal.ShuffleSlice(feedBlobs)
	for i := 0; i < pushBlobsForNumberOfRandomMessages; i++ {
		if i >= len(feedBlobs) {
			break
		}
		for _, blob := range feedBlobs[i].Blobs {
			result[blob.String()] = blob
		}
	}

	var resultSlice []refs.Blob
	for _, v := range result {
		resultSlice = append(resultSlice, v)
	}
	return resultSlice, nil
}
