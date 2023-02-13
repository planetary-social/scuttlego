package badger

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
)

type WantedFeedsRepository struct {
	graph        *SocialGraphRepository
	feedWantList *FeedWantListRepository
	feed         *FeedRepository
	banList      *BanListRepository
}

func NewWantedFeedsRepository(
	graph *SocialGraphRepository,
	feedWantList *FeedWantListRepository,
	feed *FeedRepository,
	banList *BanListRepository,
) *WantedFeedsRepository {
	return &WantedFeedsRepository{
		graph:        graph,
		feedWantList: feedWantList,
		feed:         feed,
		banList:      banList,
	}
}

func (b WantedFeedsRepository) GetWantedFeeds() (replication.WantedFeeds, error) {
	var resultContacts []replication.Contact
	var resultWantedFeeds []replication.WantedFeed

	graph, err := b.graph.GetSocialGraph()
	if err != nil {
		return replication.WantedFeeds{}, errors.Wrap(err, "could not get contacts")
	}

	for _, contact := range graph.Contacts() {
		f := contact.Id.MainFeed()

		feedState, err := b.getFeedState(f)
		if err != nil {
			return replication.WantedFeeds{}, errors.Wrap(err, "could not get feed state")
		}

		resultContacts = append(resultContacts, replication.Contact{
			Who:       f,
			Hops:      contact.Hops,
			FeedState: feedState,
		})
	}

	wantList, err := b.feedWantList.List()
	if err != nil {
		return replication.WantedFeeds{}, errors.Wrap(err, "could not get the feed want list")
	}

	for _, feedRef := range wantList {
		isBanned, err := b.banList.ContainsFeed(feedRef)
		if err != nil {
			return replication.WantedFeeds{}, errors.Wrap(err, "error checking if the feed is banned")
		}

		if isBanned {
			continue
		}

		feedState, err := b.getFeedState(feedRef)
		if err != nil {
			return replication.WantedFeeds{}, errors.Wrap(err, "could not get feed state")
		}

		resultWantedFeeds = append(resultWantedFeeds, replication.WantedFeed{
			Who:       feedRef,
			FeedState: feedState,
		})
	}

	return replication.NewWantedFeeds(resultContacts, resultWantedFeeds), nil
}

func (b WantedFeedsRepository) getFeedState(feed refs.Feed) (replication.FeedState, error) {
	f, err := b.feed.GetFeed(feed)
	if err != nil {
		if errors.Is(err, common.ErrFeedNotFound) {
			return replication.NewEmptyFeedState(), nil
		}
		return replication.FeedState{}, errors.Wrapf(err, "could not load feed '%s'", feed)
	}
	seq, ok := f.Sequence()
	if !ok {
		return replication.FeedState{}, errors.New("we got a feed so it can't be empty")
	}
	return replication.NewFeedState(seq)
}
