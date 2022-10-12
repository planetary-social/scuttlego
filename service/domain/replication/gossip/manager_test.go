package gossip_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/gossip"
	"github.com/stretchr/testify/require"
)

func TestManager_OnlyOnePeerAtATimeWillBeAskedToReplicateAFeed(t *testing.T) {
	t.Parallel()

	m := newTestManager()

	contact := replication.Contact{
		Who:       fixtures.SomeRefFeed(),
		Hops:      graph.MustNewHops(0),
		FeedState: replication.NewEmptyFeedState(),
	}

	m.Storage.Contacts = []replication.Contact{
		contact,
	}

	ctx := fixtures.TestContext(t)

	feedsCh1 := m.Manager.GetFeedsToReplicate(ctx, fixtures.SomePublicIdentity())

	select {
	case task := <-feedsCh1:
		require.Equal(t, contact.Who, task.Id)
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

	contact := replication.Contact{
		Who:       fixtures.SomeRefFeed(),
		Hops:      graph.MustNewHops(0),
		FeedState: replication.NewEmptyFeedState(),
	}

	m.Storage.Contacts = []replication.Contact{
		contact,
	}

	peer := fixtures.SomePublicIdentity()

	t.Run("first", func(t *testing.T) {
		ctx := fixtures.TestContext(t)

		select {
		case task := <-m.Manager.GetFeedsToReplicate(ctx, peer):
			require.Equal(t, contact.Who, task.Id)
			task.OnComplete(gossip.TaskResultHasMoreMessages)
		case <-time.After(1 * time.Second):
			t.Fatal("first peer should have been asked to replicate the feed")
		}
	})

	t.Run("second", func(t *testing.T) {
		ctx := fixtures.TestContext(t)

		select {
		case task := <-m.Manager.GetFeedsToReplicate(ctx, peer):
			require.Equal(t, contact.Who, task.Id)
			task.OnComplete(gossip.TaskResultDoesNotHaveMoreMessages)
		case <-time.After(1 * time.Second):
			t.Fatal("second peer should have been asked to replicate the feed")
		}
	})
}

func TestManager_TheSamePeerWillNotBeAskedForAFeedAgainRightAwayIfAllMessagesThePeerHasWereReplicated(t *testing.T) {
	t.Parallel()

	m := newTestManager()

	contact := replication.Contact{
		Who:       fixtures.SomeRefFeed(),
		Hops:      graph.MustNewHops(0),
		FeedState: replication.NewEmptyFeedState(),
	}

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
			require.Equal(t, contact.Who, task.Id)
			task.OnComplete(gossip.TaskResultDoesNotHaveMoreMessages)
		case <-time.After(1 * time.Second):
			t.Fatal("first peer should have been asked to replicate the feed")
		}
	})

	t.Run("peer2", func(t *testing.T) {
		ctx := fixtures.TestContext(t)

		select {
		case task := <-m.Manager.GetFeedsToReplicate(ctx, peer2):
			require.Equal(t, contact.Who, task.Id)
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

func TestManager_SequenceIsDeterminedBasedOnStorageStateAndMessageBufferState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name              string
		SequenceInStorage *message.Sequence
		SequenceInBuffer  *message.Sequence
		ExpectedSequence  *message.Sequence
	}{
		{
			Name:              "all_empty",
			SequenceInStorage: nil,
			SequenceInBuffer:  nil,
			ExpectedSequence:  nil,
		},
		{
			Name:              "storage_empty",
			SequenceInStorage: nil,
			SequenceInBuffer:  internal.Ptr(message.MustNewSequence(10)),
			ExpectedSequence:  internal.Ptr(message.MustNewSequence(10)),
		},
		{
			Name:              "buffer_empty",
			SequenceInStorage: internal.Ptr(message.MustNewSequence(10)),
			SequenceInBuffer:  nil,
			ExpectedSequence:  internal.Ptr(message.MustNewSequence(10)),
		},
		{
			Name:              "buffer_larger",
			SequenceInStorage: internal.Ptr(message.MustNewSequence(10)),
			SequenceInBuffer:  internal.Ptr(message.MustNewSequence(20)),
			ExpectedSequence:  internal.Ptr(message.MustNewSequence(20)),
		},
		{
			Name:              "storage_larger",
			SequenceInStorage: internal.Ptr(message.MustNewSequence(20)),
			SequenceInBuffer:  internal.Ptr(message.MustNewSequence(10)),
			ExpectedSequence:  internal.Ptr(message.MustNewSequence(20)),
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			m := newTestManager()

			contact := replication.Contact{
				Who:       fixtures.SomeRefFeed(),
				Hops:      graph.MustNewHops(0),
				FeedState: replication.NewEmptyFeedState(),
			}

			if testCase.SequenceInStorage != nil {
				contact.FeedState = replication.MustNewFeedState(*testCase.SequenceInStorage)

			} else {
				contact.FeedState = replication.NewEmptyFeedState()
			}

			m.Storage.Contacts = []replication.Contact{
				contact,
			}

			if testCase.SequenceInBuffer != nil {
				m.MessageBuffer.SetSequence(contact.Who, *testCase.SequenceInBuffer)
			}

			ctx := fixtures.TestContext(t)

			select {
			case task := <-m.Manager.GetFeedsToReplicate(ctx, fixtures.SomePublicIdentity()):
				require.Equal(t, contact.Who, task.Id)
				seq, ok := task.State.Sequence()
				if testCase.ExpectedSequence == nil {
					require.False(t, ok)
				} else {
					require.True(t, ok)
					require.Equal(t, *testCase.ExpectedSequence, seq)
				}
				task.OnComplete(gossip.TaskResultDoesNotHaveMoreMessages)
			case <-time.After(1 * time.Second):
				t.Fatal("peer should have been asked to replicate the feed")
			}
		})
	}

}

type testManager struct {
	Manager       *gossip.Manager
	Storage       *storageMock
	MessageBuffer *messageBufferMock
}

func newTestManager() testManager {
	logger := logging.NewDevNullLogger()
	storage := newStorageMock()
	buffer := newMessageBufferMock()
	manager := gossip.NewManager(logger, storage, buffer)

	return testManager{
		Manager:       manager,
		Storage:       storage,
		MessageBuffer: buffer,
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

type messageBufferMock struct {
	sequences map[string]message.Sequence
}

func newMessageBufferMock() *messageBufferMock {
	return &messageBufferMock{
		sequences: make(map[string]message.Sequence),
	}
}

func (m messageBufferMock) SetSequence(feed refs.Feed, sequence message.Sequence) {
	m.sequences[feed.String()] = sequence
}

func (m messageBufferMock) Sequence(feed refs.Feed) (message.Sequence, bool) {
	seq, ok := m.sequences[feed.String()]
	return seq, ok
}
