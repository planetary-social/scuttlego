package messagebuffer_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messagebuffer"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestFeedMessages_AddingMessagesFromTheSameFeedSucceeds(t *testing.T) {
	feed := fixtures.SomeRefFeed()

	v := messagebuffer.NewFeedMessages(feed)

	err := v.Add(fixtures.SomeTime(), someReceivedMessage(fixtures.SomeSequence(), feed))
	require.NoError(t, err)
}

func TestFeedMessages_AddingMessagesFromADifferentFeedFails(t *testing.T) {
	feed1 := fixtures.SomeRefFeed()
	feed2 := fixtures.SomeRefFeed()

	v := messagebuffer.NewFeedMessages(feed1)

	err := v.Add(fixtures.SomeTime(), someReceivedMessage(fixtures.SomeSequence(), feed2))
	require.EqualError(t, err, "incorrect feed")
}

func TestFeedMessages_LenUpdatesWhenAddingMessages(t *testing.T) {
	feed := fixtures.SomeRefFeed()

	v := messagebuffer.NewFeedMessages(feed)

	require.Equal(t, 0, v.Len())

	err := v.Add(fixtures.SomeTime(), someReceivedMessage(fixtures.SomeSequence(), feed))
	require.NoError(t, err)

	err = v.Add(fixtures.SomeTime(), someReceivedMessage(fixtures.SomeSequence(), feed))
	require.NoError(t, err)

	require.Equal(t, 2, v.Len())
}

func TestFeedMessages_RemoveOlderThanRemovesMessages(t *testing.T) {
	feed := fixtures.SomeRefFeed()

	tm := fixtures.SomeTime()

	beforeTm1 := tm.Add(-1 * time.Second)
	beforeTm2 := tm.Add(-2 * time.Second)
	beforeTm3 := tm.Add(-3 * time.Second)

	afterTm1 := tm.Add(1 * time.Second)
	afterTm2 := tm.Add(2 * time.Second)

	v := messagebuffer.NewFeedMessages(feed)

	err := v.Add(beforeTm1, someReceivedMessage(fixtures.SomeSequence(), feed))
	require.NoError(t, err)

	err = v.Add(beforeTm2, someReceivedMessage(fixtures.SomeSequence(), feed))
	require.NoError(t, err)

	err = v.Add(beforeTm3, someReceivedMessage(fixtures.SomeSequence(), feed))
	require.NoError(t, err)

	err = v.Add(afterTm1, someReceivedMessage(fixtures.SomeSequence(), feed))
	require.NoError(t, err)

	err = v.Add(afterTm2, someReceivedMessage(fixtures.SomeSequence(), feed))
	require.NoError(t, err)

	require.Equal(t, 5, v.Len())

	v.RemoveOlderThan(tm)

	require.Equal(t, 2, v.Len())
}

func TestFeedMessages_LeaveOnlyAfterDoesNothingWhenEmpty(t *testing.T) {
	feed := fixtures.SomeRefFeed()
	v := messagebuffer.NewFeedMessages(feed)

	require.Equal(t, 0, v.Len())
	v.LeaveOnlyAfter(message.MustNewSequence(2))
	require.Equal(t, 0, v.Len())
}

func TestFeedMessages_LeaveOnlyAfterDoesNotBreakWhenLastMessageIsBeingRemoved(t *testing.T) {
	feed := fixtures.SomeRefFeed()

	v := messagebuffer.NewFeedMessages(feed)

	sequence1 := message.MustNewSequence(1)
	sequence2 := message.MustNewSequence(10)
	require.GreaterOrEqual(t, sequence1.Int(), sequence1.Int())

	err := v.Add(fixtures.SomeTime(), someReceivedMessage(sequence1, feed))
	require.NoError(t, err)

	v.LeaveOnlyAfter(sequence2)
	require.Equal(t, 0, v.Len())
}

func TestFeedMessages_LeaveOnlyAfterRemovesMessages(t *testing.T) {
	feed := fixtures.SomeRefFeed()

	v := messagebuffer.NewFeedMessages(feed)

	err := v.Add(fixtures.SomeTime(), someReceivedMessage(message.MustNewSequence(1), feed))
	require.NoError(t, err)

	err = v.Add(fixtures.SomeTime(), someReceivedMessage(message.MustNewSequence(2), feed))
	require.NoError(t, err)

	err = v.Add(fixtures.SomeTime(), someReceivedMessage(message.MustNewSequence(3), feed))
	require.NoError(t, err)

	require.Equal(t, 3, v.Len())

	v.LeaveOnlyAfter(message.MustNewSequence(2))

	require.Equal(t, 1, v.Len())
	require.Equal(t,
		[]message.Sequence{
			message.MustNewSequence(3),
		},
		messagesToSequences(
			v.ConsecutiveSliceStartingWith(internal.Ptr(message.MustNewSequence(2))),
		),
	)
}

func TestFeedMessages_ConsecutiveSliceStartingWith(t *testing.T) {
	testCases := []struct {
		Name             string
		Seq              *message.Sequence
		MessageSequences []message.Sequence
		ExpectedResult   []message.Sequence
	}{
		{
			Name:             "empty",
			Seq:              nil,
			MessageSequences: nil,
			ExpectedResult:   nil,
		},
		{
			Name: "nil_consecutive",
			Seq:  nil,
			MessageSequences: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(3),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(3),
			},
		},
		{
			Name: "nil_consecutive_with_duplicates",
			Seq:  nil,
			MessageSequences: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(2),
				message.MustNewSequence(3),
				message.MustNewSequence(3),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(2),
				message.MustNewSequence(3),
				message.MustNewSequence(3),
			},
		},
		{
			Name: "nil_not_consecutive",
			Seq:  nil,
			MessageSequences: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(2),

				message.MustNewSequence(4),
				message.MustNewSequence(5),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(2),
			},
		},
		{
			Name: "nil_not_consecutive_with_duplicates",
			Seq:  nil,
			MessageSequences: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(2),

				message.MustNewSequence(4),
				message.MustNewSequence(4),
				message.MustNewSequence(5),
				message.MustNewSequence(5),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(2),
			},
		},
		{
			Name: "nil_gap_at_the_start",
			Seq:  nil,
			MessageSequences: []message.Sequence{
				message.MustNewSequence(2),
				message.MustNewSequence(3),
				message.MustNewSequence(4),
			},
			ExpectedResult: nil,
		},

		{
			Name: "not_nil_consecutive",
			Seq:  internal.Ptr(message.MustNewSequence(2)),
			MessageSequences: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(5),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(5),
			},
		},
		{
			Name: "not_nil_consecutive_with_duplicates",
			Seq:  internal.Ptr(message.MustNewSequence(2)),
			MessageSequences: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(4),
				message.MustNewSequence(5),
				message.MustNewSequence(5),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(4),
				message.MustNewSequence(5),
				message.MustNewSequence(5),
			},
		},
		{
			Name: "not_nil_not_consecutive",
			Seq:  internal.Ptr(message.MustNewSequence(2)),
			MessageSequences: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(4),

				message.MustNewSequence(6),
				message.MustNewSequence(7),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(4),
			},
		},
		{
			Name: "not_nil_not_consecutive_with_duplicates",
			Seq:  internal.Ptr(message.MustNewSequence(2)),
			MessageSequences: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(4),

				message.MustNewSequence(6),
				message.MustNewSequence(6),
				message.MustNewSequence(7),
				message.MustNewSequence(7),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(4),
			},
		},
		{
			Name: "nil_gap_at_the_start",
			Seq:  internal.Ptr(message.MustNewSequence(2)),
			MessageSequences: []message.Sequence{
				message.MustNewSequence(4),
				message.MustNewSequence(5),
				message.MustNewSequence(6),
			},
			ExpectedResult: nil,
		},

		{
			Name: "discard_starting_not_nil_consecutive",
			Seq:  internal.Ptr(message.MustNewSequence(2)),
			MessageSequences: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(5),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(5),
			},
		},
		{
			Name: "discard_starting_not_nil_consecutive_with_duplicates",
			Seq:  internal.Ptr(message.MustNewSequence(2)),
			MessageSequences: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(2),
				message.MustNewSequence(3),
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(4),
				message.MustNewSequence(5),
				message.MustNewSequence(5),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(4),
				message.MustNewSequence(5),
				message.MustNewSequence(5),
			},
		},
		{
			Name: "discard_starting_not_nil_not_consecutive",
			Seq:  internal.Ptr(message.MustNewSequence(2)),
			MessageSequences: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(3),
				message.MustNewSequence(4),

				message.MustNewSequence(6),
				message.MustNewSequence(7),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(4),
			},
		},
		{
			Name: "discard_starting_not_nil_not_consecutive_with_duplicates",
			Seq:  internal.Ptr(message.MustNewSequence(2)),
			MessageSequences: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(2),
				message.MustNewSequence(3),
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(4),

				message.MustNewSequence(6),
				message.MustNewSequence(6),
				message.MustNewSequence(7),
				message.MustNewSequence(7),
			},
			ExpectedResult: []message.Sequence{
				message.MustNewSequence(3),
				message.MustNewSequence(3),
				message.MustNewSequence(4),
				message.MustNewSequence(4),
			},
		},
		{
			Name: "discard_starting_nil_gap_at_the_start",
			Seq:  internal.Ptr(message.MustNewSequence(2)),
			MessageSequences: []message.Sequence{
				message.MustNewSequence(1),
				message.MustNewSequence(2),
				message.MustNewSequence(4),
				message.MustNewSequence(5),
				message.MustNewSequence(6),
			},
			ExpectedResult: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			feed := fixtures.SomeRefFeed()
			v := messagebuffer.NewFeedMessages(feed)

			for _, seq := range testCase.MessageSequences {
				err := v.Add(fixtures.SomeTime(), someReceivedMessage(seq, feed))
				require.NoError(t, err)
			}

			msgs := v.ConsecutiveSliceStartingWith(testCase.Seq)
			require.Equal(t, testCase.ExpectedResult, messagesToSequences(msgs))
		})
	}

}

func TestFeedMessages_RemoveRemovesMessages(t *testing.T) {
	feed := fixtures.SomeRefFeed()

	v := messagebuffer.NewFeedMessages(feed)

	msg1 := someReceivedMessage(fixtures.SomeSequence(), feed)
	msg2 := someReceivedMessage(fixtures.SomeSequence(), feed)

	err := v.Add(fixtures.SomeTime(), msg1)
	require.NoError(t, err)

	err = v.Add(fixtures.SomeTime(), msg2)
	require.NoError(t, err)

	require.Equal(t, 2, v.Len())
	v.Remove(msg2.Message().Raw())
	require.Equal(t, 1, v.Len())
}

func messagesToSequences(msgs []messagebuffer.ReceivedMessage[feeds.PeekedMessage]) []message.Sequence {
	var result []message.Sequence
	for _, msg := range msgs {
		result = append(result, msg.Message().Sequence())
	}
	return result
}

func someReceivedMessage(seq message.Sequence, feed refs.Feed) messagebuffer.ReceivedMessage[feeds.PeekedMessage] {
	return messagebuffer.MustNewReceivedMessage(
		fixtures.SomePublicIdentity(),
		feeds.MustNewPeekedMessage(
			feed,
			seq,
			fixtures.SomeRawMessage(),
		),
	)
}
