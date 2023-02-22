package ebt_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/stretchr/testify/require"
)

func TestSentNotes_UpdateReturnsNotesForNewContacts(t *testing.T) {
	sn := ebt.NewSentNotes()

	contact1Seq := fixtures.SomeSequence()
	contact1 := replication.MustNewContact(
		fixtures.SomeRefFeed(),
		graph.MustNewHops(1),
		replication.MustNewFeedState(contact1Seq),
	)

	contact2 := replication.MustNewContact(
		fixtures.SomeRefFeed(),
		graph.MustNewHops(2),
		replication.NewEmptyFeedState(),
	)

	contacts := []replication.Contact{
		contact1,
		contact2,
	}

	notes, err := sn.Update(contacts)
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNote{
			messages.MustNewEbtReplicateNote(
				contact1.Who(),
				true,
				true,
				contact1Seq.Int(),
			),
			messages.MustNewEbtReplicateNote(
				contact2.Who(),
				true,
				true,
				0,
			),
		},
		notes.Notes(),
	)

}

func TestSentNotes_UpdateReturnsUpdatesForOldContactsWithNewerSequenceNumbers(t *testing.T) {
	sn := ebt.NewSentNotes()

	contactRef := fixtures.SomeRefFeed()
	contactSeq1 := message.MustNewSequence(9)
	contactSeq2 := message.MustNewSequence(10)

	notes, err := sn.Update([]replication.Contact{
		replication.MustNewContact(
			contactRef,
			graph.MustNewHops(1),
			replication.MustNewFeedState(contactSeq1),
		),
	})
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNote{
			messages.MustNewEbtReplicateNote(
				contactRef,
				true,
				true,
				contactSeq1.Int(),
			),
		},
		notes.Notes(),
	)

	notes, err = sn.Update([]replication.Contact{
		replication.MustNewContact(
			contactRef,
			graph.MustNewHops(1),
			replication.MustNewFeedState(contactSeq2),
		),
	})
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNote{
			messages.MustNewEbtReplicateNote(
				contactRef,
				true,
				true,
				contactSeq2.Int(),
			),
		},
		notes.Notes(),
	)
}

func TestSentNotes_UpdateDoesNotReturnUpdatesForOldContactsWithIdenticalSequenceNumbers(t *testing.T) {
	sn := ebt.NewSentNotes()

	contactRef := fixtures.SomeRefFeed()
	contactSeq := message.MustNewSequence(10)

	contacts := []replication.Contact{
		replication.MustNewContact(
			contactRef,
			graph.MustNewHops(1),
			replication.MustNewFeedState(contactSeq),
		),
	}

	notes, err := sn.Update(contacts)
	require.NoError(t, err)
	require.Equal(t,
		[]messages.EbtReplicateNote{
			messages.MustNewEbtReplicateNote(
				contactRef,
				true,
				true,
				contactSeq.Int(),
			),
		},
		notes.Notes(),
	)

	notes, err = sn.Update(contacts)
	require.NoError(t, err)
	require.Empty(t, notes.Notes())
}

func TestSentNotes_UpdateReturnsUpdatesForOldContactsWithOlderSequenceNumbers(t *testing.T) {
	sn := ebt.NewSentNotes()

	contactRef := fixtures.SomeRefFeed()
	contactSeq1 := message.MustNewSequence(10)
	contactSeq2 := message.MustNewSequence(9)

	notes, err := sn.Update([]replication.Contact{
		replication.MustNewContact(
			contactRef,
			graph.MustNewHops(1),
			replication.MustNewFeedState(contactSeq1),
		),
	})
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNote{
			messages.MustNewEbtReplicateNote(
				contactRef,
				true,
				true,
				contactSeq1.Int(),
			),
		},
		notes.Notes(),
	)

	notes, err = sn.Update([]replication.Contact{
		replication.MustNewContact(
			contactRef,
			graph.MustNewHops(1),
			replication.MustNewFeedState(contactSeq2),
		),
	})
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNote{
			messages.MustNewEbtReplicateNote(
				contactRef,
				true,
				true,
				contactSeq2.Int(),
			),
		},
		notes.Notes(),
	)
}

func TestSentNotes_UpdateReturnsCancellationUpdatesForMissingContactsAndSendsNotesAgainIfContactReappears(t *testing.T) {
	sn := ebt.NewSentNotes()

	contact1 := replication.MustNewContact(
		fixtures.SomeRefFeed(),
		graph.MustNewHops(1),
		replication.NewEmptyFeedState(),
	)

	contact2 := replication.MustNewContact(
		fixtures.SomeRefFeed(),
		graph.MustNewHops(2),
		replication.NewEmptyFeedState(),
	)

	notes, err := sn.Update([]replication.Contact{
		contact1,
		contact2,
	})
	require.NoError(t, err)
	require.Equal(t,
		[]messages.EbtReplicateNote{
			messages.MustNewEbtReplicateNote(
				contact1.Who(),
				true,
				true,
				0,
			),
			messages.MustNewEbtReplicateNote(
				contact2.Who(),
				true,
				true,
				0,
			),
		},
		notes.Notes(),
	)

	notes, err = sn.Update([]replication.Contact{
		contact1,
	})
	require.NoError(t, err)
	require.Equal(t,
		[]messages.EbtReplicateNote{
			messages.MustNewEbtReplicateNote(
				contact2.Who(),
				false,
				false,
				-1,
			),
		},
		notes.Notes(),
	)

	notes, err = sn.Update([]replication.Contact{
		contact1,
		contact2,
	})
	require.NoError(t, err)
	require.Equal(t,
		[]messages.EbtReplicateNote{
			messages.MustNewEbtReplicateNote(
				contact2.Who(),
				true,
				true,
				0,
			),
		},
		notes.Notes(),
	)
}
