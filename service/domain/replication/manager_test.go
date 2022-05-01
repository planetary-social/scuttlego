package replication_test

import (
	"testing"
	"time"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/domain/graph"
	"github.com/planetary-social/go-ssb/service/domain/replication"
	"github.com/stretchr/testify/require"
)

func TestManager_OnlyOnePeerAtATimeWillBeAskedToReplicateAFeed(t *testing.T) {
	t.Parallel()

	logger := fixtures.TestLogger(t)
	storage := newStorageMock()
	manager := replication.NewManager(logger, storage)

	contact := replication.Contact{
		Who:       fixtures.SomeRefFeed(),
		Hops:      graph.MustNewHops(0),
		FeedState: replication.NewEmptyFeedState(),
	}

	storage.Contacts = []replication.Contact{
		contact,
	}

	ctx := fixtures.TestContext(t)

	feedsCh1 := manager.GetFeedsToReplicate(ctx, fixtures.SomePublicIdentity())

	select {
	case task := <-feedsCh1:
		require.Equal(t, contact.Who, task.Id)
	case <-time.After(1 * time.Second):
		t.Fatal("first peer should have been asked to replicate the feed")
	}

	feedsCh2 := manager.GetFeedsToReplicate(ctx, fixtures.SomePublicIdentity())

	select {
	case <-feedsCh2:
		t.Fatal("second peer should not replicate the feed at the same time")
	case <-time.After(1 * time.Second):
		// correct, nothing received
	}
}

type storageMock struct {
	Contacts []replication.Contact
}

func newStorageMock() *storageMock {
	return &storageMock{}
}

func (s storageMock) GetContacts() ([]replication.Contact, error) {
	return s.Contacts, nil
}
