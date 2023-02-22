package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
)

type WantedFeedsProvider struct {
	transaction TransactionProvider
}

func NewWantedFeedsProvider(transaction TransactionProvider) *WantedFeedsProvider {
	return &WantedFeedsProvider{transaction: transaction}
}

func (b *WantedFeedsProvider) GetWantedFeeds() (replication.WantedFeeds, error) {
	var resultContacts []replication.Contact
	var resultWantedFeeds []replication.WantedFeed

	if err := b.transaction.Transact(func(adapters Adapters) error {
		graph, err := adapters.SocialGraph.GetSocialGraph()
		if err != nil {
			return errors.Wrap(err, "could not get the social graph")
		}

		for _, graphContact := range graph.Contacts() {
			feedRef := graphContact.Id.MainFeed()

			feedState, err := b.getFeedState(adapters, feedRef)
			if err != nil {
				return errors.Wrap(err, "could not get contact feed state")
			}

			contact, err := replication.NewContact(feedRef, graphContact.Hops, feedState)
			if err != nil {
				return errors.Wrap(err, "error creating a contact")
			}

			resultContacts = append(resultContacts, contact)
		}

		wantList, err := adapters.FeedWantList.List()
		if err != nil {
			return errors.Wrap(err, "could not get the feed want list")
		}

		for _, feedRef := range wantList {
			isBanned, err := adapters.BanList.ContainsFeed(feedRef)
			if err != nil {
				return errors.Wrap(err, "error checking if the feed is banned")
			}

			if isBanned {
				continue
			}

			feedState, err := b.getFeedState(adapters, feedRef)
			if err != nil {
				return errors.Wrap(err, "could not get wanted feed feed state")
			}

			wantedFeed, err := replication.NewWantedFeed(feedRef, feedState)
			if err != nil {
				return errors.Wrap(err, "error creating a wanted feed")
			}

			resultWantedFeeds = append(resultWantedFeeds, wantedFeed)
		}

		return nil
	}); err != nil {
		return replication.WantedFeeds{}, errors.Wrap(err, "transaction error")
	}

	return replication.NewWantedFeeds(resultContacts, resultWantedFeeds)
}

func (b *WantedFeedsProvider) getFeedState(adapters Adapters, feed refs.Feed) (replication.FeedState, error) {
	f, err := adapters.Feed.GetFeed(feed)
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
