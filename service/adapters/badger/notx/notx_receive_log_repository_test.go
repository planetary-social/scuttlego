package notx_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
)

func TestNoTxReceiveLogRepository_GetSequences(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	msgRef := fixtures.SomeRefMessage()

	_, err := ts.NoTxTestAdapters.NoTxReceiveLogRepository.GetSequences(msgRef)
	require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.ReceiveLogRepository.Put(msgRef)
	})
	require.NoError(t, err)

	sequences, err := ts.NoTxTestAdapters.NoTxReceiveLogRepository.GetSequences(msgRef)
	require.NoError(t, err)
	require.NotEmpty(t, sequences)
}

func TestNoTxReceiveLogRepository_GetMessage(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	msg := fixtures.SomeMessageWithUniqueRawMessage(message.NewFirstSequence(), fixtures.SomeRefFeed())
	seq := fixtures.SomeReceiveLogSequence()

	ts.Dependencies.RawMessageIdentifier.Mock(msg)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.ReceiveLogRepository.PutUnderSpecificSequence(msg.Id(), seq)
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.MessageRepository.Put(msg)
	})
	require.NoError(t, err)

	retrievedMsg, err := ts.NoTxTestAdapters.NoTxReceiveLogRepository.GetMessage(seq)
	require.NoError(t, err)
	require.Equal(t, msg, retrievedMsg)
}

func TestNoTxReceiveLogRepository_List(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	msg := fixtures.SomeMessageWithUniqueRawMessage(message.NewFirstSequence(), fixtures.SomeRefFeed())
	seq := fixtures.SomeReceiveLogSequence()

	ts.Dependencies.RawMessageIdentifier.Mock(msg)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.ReceiveLogRepository.PutUnderSpecificSequence(msg.Id(), seq)
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.MessageRepository.Put(msg)
	})
	require.NoError(t, err)

	msgs, err := ts.NoTxTestAdapters.NoTxReceiveLogRepository.List(common.MustNewReceiveLogSequence(0), 10)
	require.NoError(t, err)
	require.Equal(t, []queries.LogMessage{
		{
			Message:  msg,
			Sequence: seq,
		},
	}, msgs)
}
