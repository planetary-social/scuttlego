package badger_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/stretchr/testify/require"
)

func TestReceiveLog_GetMessage_ReturnsPredefinedErrorWhenNotFound(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		sequence1 := fixtures.SomeReceiveLogSequence()
		sequence2 := fixtures.SomeReceiveLogSequence()

		_, err := adapters.ReceiveLogRepository.GetMessage(sequence1)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		_, err = adapters.ReceiveLogRepository.GetMessage(sequence2)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		err = adapters.ReceiveLogRepository.PutUnderSpecificSequence(msg.Id(), sequence1)
		require.NoError(t, err)

		err = adapters.MessageRepository.Put(msg)
		require.NoError(t, err)

		_, err = adapters.ReceiveLogRepository.GetMessage(sequence1)
		require.NoError(t, err)

		_, err = adapters.ReceiveLogRepository.GetMessage(sequence2)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_GetSequences_ReturnsPredefinedErrorWhenNotFound(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msg1 := fixtures.SomeRefMessage()
	msg2 := fixtures.SomeRefMessage()

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		_, err := adapters.ReceiveLogRepository.GetSequences(msg1)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		_, err = adapters.ReceiveLogRepository.GetSequences(msg2)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		err = adapters.ReceiveLogRepository.PutUnderSpecificSequence(msg1, fixtures.SomeReceiveLogSequence())
		require.NoError(t, err)

		_, err = adapters.ReceiveLogRepository.GetSequences(msg1)
		require.NoError(t, err)

		_, err = adapters.ReceiveLogRepository.GetSequences(msg2)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLogRepository_CallingPutMultipleTimesInsertsMessageMultipleTimes(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		if err := adapters.ReceiveLogRepository.Put(msg.Id()); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := adapters.ReceiveLogRepository.Put(msg.Id()); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		seqs, err := adapters.ReceiveLogRepository.GetSequences(msg.Id())
		require.NoError(t, err)

		require.Equal(t,
			[]common.ReceiveLogSequence{
				common.MustNewReceiveLogSequence(0),
				common.MustNewReceiveLogSequence(1),
			},
			seqs)

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Get_ReturnsNoMessagesWhenEmpty(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		msgs, err := adapters.ReceiveLogRepository.List(common.MustNewReceiveLogSequence(0), 10)
		require.NoError(t, err)
		require.Empty(t, msgs)

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Get_ReturnsErrorForInvalidLimit(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	err := ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		_, err := adapters.ReceiveLogRepository.List(common.MustNewReceiveLogSequence(0), 0)
		require.EqualError(t, err, "limit must be positive")

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Put_InsertsCorrectMapping(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	expectedSequence := common.MustNewReceiveLogSequence(0)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		if err := adapters.ReceiveLogRepository.Put(msg.Id()); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := adapters.MessageRepository.Put(msg); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		seqs, err := adapters.ReceiveLogRepository.GetSequences(msg.Id())
		require.NoError(t, err)
		require.Equal(t, []common.ReceiveLogSequence{expectedSequence}, seqs)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		_, err = adapters.ReceiveLogRepository.GetMessage(expectedSequence)
		require.NoError(t, err)
		// retrieved message won't have the same fields as the message we saved
		// as the raw data set in fixtures.SomeMessage is gibberish

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Get_ReturnsMessagesObeyingLimitAndStartSeq(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	numMessages := 10

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		for i := 0; i < numMessages; i++ {
			msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

			if err := adapters.ReceiveLogRepository.Put(msg.Id()); err != nil {
				return errors.Wrap(err, "could not put a message in receive log")
			}

			if err := adapters.MessageRepository.Put(msg); err != nil {
				return errors.Wrap(err, "could not put a message in message repository")
			}
		}

		return nil
	})
	require.NoError(t, err)

	t.Run("seq_0", func(t *testing.T) {
		err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
			msgs, err := adapters.ReceiveLogRepository.List(common.MustNewReceiveLogSequence(0), 10)
			require.NoError(t, err)
			require.Len(t, msgs, 10)

			return nil
		})
		require.NoError(t, err)
	})

	t.Run("seq_5", func(t *testing.T) {
		err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
			msgs, err := adapters.ReceiveLogRepository.List(common.MustNewReceiveLogSequence(5), 10)
			require.NoError(t, err)
			require.Len(t, msgs, 5)

			return nil
		})
		require.NoError(t, err)
	})
}

func TestReceiveLog_PutUnderSpecificSequence_InsertsCorrectMapping(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	sequence := common.MustNewReceiveLogSequence(123)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		if err := adapters.ReceiveLogRepository.PutUnderSpecificSequence(msg.Id(), sequence); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := adapters.MessageRepository.Put(msg); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		seqs, err := adapters.ReceiveLogRepository.GetSequences(msg.Id())
		require.NoError(t, err)
		require.Equal(t, []common.ReceiveLogSequence{sequence}, seqs)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		_, err = adapters.ReceiveLogRepository.GetMessage(sequence)
		require.NoError(t, err)
		// retrieved message won't have the same fields as the message we saved
		// as the raw data set in fixtures.SomeMessage is gibberish

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLogRepository_PutUnderSpecificSequenceAdvancesInternalSequenceCounterIfItIsLower(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msg1 := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	sequence := common.MustNewReceiveLogSequence(123)

	msg2 := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		if err := adapters.ReceiveLogRepository.PutUnderSpecificSequence(msg1.Id(), sequence); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := adapters.MessageRepository.Put(msg1); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		msgs, err := adapters.ReceiveLogRepository.List(common.MustNewReceiveLogSequence(0), 100)
		require.NoError(t, err)

		require.Len(t, msgs, 1)
		require.Equal(t, sequence, msgs[0].Sequence)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		if err := adapters.ReceiveLogRepository.Put(msg2.Id()); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := adapters.MessageRepository.Put(msg2); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		msgs, err := adapters.ReceiveLogRepository.List(common.MustNewReceiveLogSequence(0), 100)
		require.NoError(t, err)

		require.Len(t, msgs, 2)
		require.Equal(t, sequence, msgs[0].Sequence)
		require.Equal(t, common.MustNewReceiveLogSequence(sequence.Int()+1), msgs[1].Sequence)

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLogRepository_PutUnderSpecificSequenceDoesNotAdvanceInternalSequenceCounterIfItIsHigher(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msg1 := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	msg2 := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	msg3 := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	sequence := common.MustNewReceiveLogSequence(0)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		if err := adapters.ReceiveLogRepository.Put(msg1.Id()); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := adapters.MessageRepository.Put(msg1); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		if err := adapters.ReceiveLogRepository.Put(msg2.Id()); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := adapters.MessageRepository.Put(msg2); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		msgs, err := adapters.ReceiveLogRepository.List(common.MustNewReceiveLogSequence(0), 100)
		require.NoError(t, err)

		require.Len(t, msgs, 2)
		require.Equal(t, common.MustNewReceiveLogSequence(0), msgs[0].Sequence)
		require.Equal(t, common.MustNewReceiveLogSequence(1), msgs[1].Sequence)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		if err := adapters.ReceiveLogRepository.PutUnderSpecificSequence(msg1.Id(), sequence); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		if err := adapters.ReceiveLogRepository.Put(msg3.Id()); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := adapters.MessageRepository.Put(msg3); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		msgs, err := adapters.ReceiveLogRepository.List(common.MustNewReceiveLogSequence(0), 100)
		require.NoError(t, err)

		require.Len(t, msgs, 3)
		require.Equal(t, common.MustNewReceiveLogSequence(0), msgs[0].Sequence)
		require.Equal(t, common.MustNewReceiveLogSequence(1), msgs[1].Sequence)
		require.Equal(t, common.MustNewReceiveLogSequence(2), msgs[2].Sequence)

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLogRepository_OneMessageMayBeUnderMultipleSequences(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	sequence1 := common.MustNewReceiveLogSequence(123)
	sequence2 := common.MustNewReceiveLogSequence(345)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		if err := adapters.ReceiveLogRepository.PutUnderSpecificSequence(msg.Id(), sequence1); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := adapters.ReceiveLogRepository.PutUnderSpecificSequence(msg.Id(), sequence2); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := adapters.MessageRepository.Put(msg); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		seqs, err := adapters.ReceiveLogRepository.GetSequences(msg.Id())
		require.NoError(t, err)

		require.Equal(t,
			[]common.ReceiveLogSequence{
				sequence1,
				sequence2,
			},
			seqs)

		return nil
	})
	require.NoError(t, err)
}
