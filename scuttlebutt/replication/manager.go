package replication

import (
	"context"
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds"
)

var ErrFeedNotFound = errors.New("feed not found")

type FeedStorage interface {
	GetFeed(ref refs.Feed) (*feeds.Feed, error)
}

type Manager struct {
	//feedStorage FeedStorage
	logger logging.Logger
}

func NewManager(logger logging.Logger /*, feedStorage FeedStorage*/) *Manager {
	return &Manager{
		//feedStorage: feedStorage,
		logger: logger.New("manager"),
	}
}

func (m Manager) GetFeedsToReplicate(ctx context.Context) (<-chan ReplicateFeedTask, error) {
	ch := make(chan ReplicateFeedTask)

	go func() {
		for _, ref := range m.getFeedsToReplicate() {
			feedState, err := m.getFeedState(ref)
			if err != nil {
				panic(err) // todo do not panic
			}

			ch <- ReplicateFeedTask{
				Id:    ref,
				State: feedState,
				Ctx:   ctx,
			}
		}
	}()

	return ch, nil
}

func (m Manager) getFeedState(ref refs.Feed) (FeedState, error) {
	//feed, err := m.feedStorage.GetFeed(ref)
	//if err != nil {
	//	if errors.Is(err, ErrFeedNotFound) {
	//		return NewEmptyFeedState(), nil
	//	}
	//	return FeedState{}, errors.Wrap(err, "could not get a feed")
	//}
	return NewEmptyFeedState(), nil
}

func (m Manager) getFeedsToReplicate() []refs.Feed {
	return []refs.Feed{
		refs.MustNewFeed("@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519"),
	}
}
