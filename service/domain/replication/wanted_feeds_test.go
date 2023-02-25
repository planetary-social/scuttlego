package replication_test

import (
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"testing"
)

func BenchmarkWantedFeedsCache_GetContacts(b *testing.B) {
	wantedFeedsProvider := newWantedFeedsProviderMock()

	var contacts []replication.Contact
	for i := 0; i < 5000; i++ {
		contacts = append(contacts,
			replication.MustNewContact(
				fixtures.SomeRefFeed(),
				fixtures.SomeHops(),
				replication.NewEmptyFeedState(),
			),
		)
	}

	wantedFeedsProvider.GetWantedFeedsReturnValue = replication.MustNewWantedFeeds(contacts, nil)

	peer := fixtures.SomePublicIdentity()
	cache := replication.NewWantedFeedsCache(wantedFeedsProvider)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := cache.GetContacts(peer)
		if err != nil {
			b.Fatal(err)
		}
	}
}

type wantedFeedsProviderMock struct {
	GetWantedFeedsReturnValue replication.WantedFeeds
}

func newWantedFeedsProviderMock() *wantedFeedsProviderMock {
	return &wantedFeedsProviderMock{}
}

func (w wantedFeedsProviderMock) GetWantedFeeds() (replication.WantedFeeds, error) {
	return w.GetWantedFeedsReturnValue, nil
}
