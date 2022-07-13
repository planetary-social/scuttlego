package bolt_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestReceiveLog_Get_ReturnsNoMessagesWhenEmpty(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		msgs, err := txadapters.ReceiveLog.List(queries.MustNewReceiveLogSequence(0), 10)
		require.NoError(t, err)
		require.Empty(t, msgs)

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Get_ReturnsErrorForInvalidLimit(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = txadapters.ReceiveLog.List(queries.MustNewReceiveLogSequence(0), 0)
		require.EqualError(t, err, "limit must be positive")

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Put_InsertsCorrectMapping(t *testing.T) {
	db := fixtures.Bolt(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	expectedSequence := queries.MustNewReceiveLogSequence(0)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		if err := txadapters.ReceiveLog.Put(msg.Id()); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := txadapters.MessageRepository.Put(msg); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		seq, err := txadapters.ReceiveLog.GetSequence(msg.Id())
		require.NoError(t, err)
		require.Equal(t, expectedSequence, seq)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = txadapters.ReceiveLog.GetMessage(expectedSequence)
		require.NoError(t, err)
		// retrieved message won't have the same fields as the message we saved
		// as the raw data set in fixtures.SomeMessage is gibberish

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Get_ReturnsMessagesObeyingLimitAndStartSeq(t *testing.T) {
	db := fixtures.Bolt(t)

	numMessages := 10

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		for i := 0; i < numMessages; i++ {
			msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

			if err := txadapters.ReceiveLog.Put(msg.Id()); err != nil {
				return errors.Wrap(err, "could not put a message in receive log")
			}

			if err := txadapters.MessageRepository.Put(msg); err != nil {
				return errors.Wrap(err, "could not put a message in message repository")
			}
		}

		return nil
	})
	require.NoError(t, err)

	t.Run("seq_0", func(t *testing.T) {
		err = db.Update(func(tx *bbolt.Tx) error {
			txadapters, err := di.BuildTxTestAdapters(tx)
			require.NoError(t, err)

			msgs, err := txadapters.ReceiveLog.List(queries.MustNewReceiveLogSequence(0), 10)
			require.NoError(t, err)
			require.Len(t, msgs, 10)

			return nil
		})
		require.NoError(t, err)
	})

	t.Run("seq_5", func(t *testing.T) {
		err = db.Update(func(tx *bbolt.Tx) error {
			txadapters, err := di.BuildTxTestAdapters(tx)
			require.NoError(t, err)

			msgs, err := txadapters.ReceiveLog.List(queries.MustNewReceiveLogSequence(5), 10)
			require.NoError(t, err)
			require.Len(t, msgs, 5)

			return nil
		})
		require.NoError(t, err)
	})
}
