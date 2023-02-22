package gossip_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/gossip"
	"github.com/stretchr/testify/require"
)

func TestManager_OnlyOnePeerAtATimeWillBeAskedToReplicateAFeed(t *testing.T) {
	t.Parallel()

	m := newTestManager()

	contact := replication.MustNewContact(
		fixtures.SomeRefFeed(),
		graph.MustNewHops(0),
		replication.NewEmptyFeedState(),
	)

	m.Storage.Contacts = []replication.Contact{
		contact,
	}

	ctx := fixtures.TestContext(t)

	feedsCh1 := m.Manager.GetFeedsToReplicate(ctx, fixtures.SomePublicIdentity())

	select {
	case task := <-feedsCh1:
		require.Equal(t, contact.Who(), task.Id)
	case <-time.After(1 * time.Second):
		t.Fatal("first peer should have been asked to replicate the feed")
	}

	feedsCh2 := m.Manager.GetFeedsToReplicate(ctx, fixtures.SomePublicIdentity())

	select {
	case <-feedsCh2:
		t.Fatal("second peer should not replicate the feed at the same time")
	case <-time.After(1 * time.Second):
		// correct, nothing received
	}
}

func TestManager_TheSamePeerWillBeAskedForAFeedAgainRightAwayIfNotAllMessagesThePeerHasWereReplicated(t *testing.T) {
	t.Parallel()

	m := newTestManager()

	contact := replication.MustNewContact(
		fixtures.SomeRefFeed(),
		graph.MustNewHops(0),
		replication.NewEmptyFeedState(),
	)

	m.Storage.Contacts = []replication.Contact{
		contact,
	}

	peer := fixtures.SomePublicIdentity()

	t.Run("first", func(t *testing.T) {
		ctx := fixtures.TestContext(t)

		select {
		case task := <-m.Manager.GetFeedsToReplicate(ctx, peer):
			require.Equal(t, contact.Who(), task.Id)
			task.OnComplete(gossip.TaskResultHasMoreMessages)
		case <-time.After(1 * time.Second):
			t.Fatal("first peer should have been asked to replicate the feed")
		}
	})

	t.Run("second", func(t *testing.T) {
		ctx := fixtures.TestContext(t)

		select {
		case task := <-m.Manager.GetFeedsToReplicate(ctx, peer):
			require.Equal(t, contact.Who(), task.Id)
			task.OnComplete(gossip.TaskResultDoesNotHaveMoreMessages)
		case <-time.After(1 * time.Second):
			t.Fatal("second peer should have been asked to replicate the feed")
		}
	})
}

func TestManager_TheSamePeerWillNotBeAskedForAFeedAgainRightAwayIfAllMessagesThePeerHasWereReplicated(t *testing.T) {
	t.Parallel()

	m := newTestManager()

	contact := replication.MustNewContact(
		fixtures.SomeRefFeed(),
		graph.MustNewHops(0),
		replication.NewEmptyFeedState(),
	)

	m.Storage.Contacts = []replication.Contact{
		contact,
	}

	peer1 := fixtures.SomePublicIdentity()
	peer2 := fixtures.SomePublicIdentity()

	// separate subtests are used to make controlling context terminations easier

	t.Run("peer1_first", func(t *testing.T) {
		ctx := fixtures.TestContext(t)

		select {
		case task := <-m.Manager.GetFeedsToReplicate(ctx, peer1):
			require.Equal(t, contact.Who(), task.Id)
			task.OnComplete(gossip.TaskResultDoesNotHaveMoreMessages)
		case <-time.After(1 * time.Second):
			t.Fatal("first peer should have been asked to replicate the feed")
		}
	})

	t.Run("peer2", func(t *testing.T) {
		ctx := fixtures.TestContext(t)

		select {
		case task := <-m.Manager.GetFeedsToReplicate(ctx, peer2):
			require.Equal(t, contact.Who(), task.Id)
			task.OnComplete(gossip.TaskResultDoesNotHaveMoreMessages)
		case <-time.After(1 * time.Second):
			t.Fatal("second peer should have been asked to replicate the feed")
		}
	})

	t.Run("peer1_second", func(t *testing.T) {
		ctx := fixtures.TestContext(t)

		select {
		case <-m.Manager.GetFeedsToReplicate(ctx, peer1):
			t.Fatal("first peer should not replicate the feed again as it is too early (backoff)")
		case <-time.After(1 * time.Second):
			// correct, nothing received
		}
	})
}

func TestManager_OnlyLocalFeedWillBeSentFromGetFeedsToReplicateSelf(t *testing.T) {
	t.Parallel()

	m := newTestManager()

	contact1 := replication.MustNewContact(
		fixtures.SomeRefFeed(),
		graph.MustNewHops(0),
		replication.NewEmptyFeedState(),
	)

	contact2 := replication.MustNewContact(
		fixtures.SomeRefFeed(),
		graph.MustNewHops(fixtures.SomePositiveInt()),
		replication.NewEmptyFeedState(),
	)

	m.Storage.Contacts = []replication.Contact{
		contact1,
		contact2,
	}

	ctx := fixtures.TestContext(t)

	feedsCh := m.Manager.GetFeedsToReplicateSelf(ctx, fixtures.SomePublicIdentity())

	select {
	case task := <-feedsCh:
		require.Equal(t, contact1.Who(), task.Id)
	case <-time.After(1 * time.Second):
		t.Fatal("peer should have been asked to replicate the local feed")
	}

	select {
	case <-feedsCh:
		t.Fatal("peer should not replicate other feeds")
	case <-time.After(1 * time.Second):
		// correct, nothing received
	}
}

type testManager struct {
	Manager *gossip.Manager
	Storage *storageMock
}

func newTestManager() testManager {
	logger := logging.NewDevNullLogger()
	storage := newStorageMock()
	manager := gossip.NewManager(logger, storage)

	return testManager{
		Manager: manager,
		Storage: storage,
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
