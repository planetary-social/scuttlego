package bolt_test

import (
	"github.com/planetary-social/go-ssb/cmd/ssb-test/di"
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/adapters/bolt"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestMessageRepository_CountEmpty(t *testing.T) {
	db := fixtures.Bolt(t)
	identifier := NewRawMessageIdentifierMock()

	err := db.View(func(tx *bbolt.Tx) error {
		repository := bolt.NewMessageRepository(tx, identifier)

		n, err := repository.Count()
		require.NoError(t, err)
		require.Equal(t, 0, n)

		return nil
	})
	require.NoError(t, err)
}

func TestMessageRepository_GetNoMessage(t *testing.T) {
	db := fixtures.Bolt(t)
	identifier := NewRawMessageIdentifierMock()

	err := db.View(func(tx *bbolt.Tx) error {
		repository := bolt.NewMessageRepository(tx, identifier)

		_, err := repository.Get(fixtures.SomeRefMessage())
		require.EqualError(t, err, "message not found")

		return nil
	})
	require.NoError(t, err)
}

func TestMessageRepository_Put_Get(t *testing.T) {
	db := fixtures.Bolt(t)
	identifier := NewRawMessageIdentifierMock()

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := db.Update(func(tx *bbolt.Tx) error {
		repository := bolt.NewMessageRepository(tx, identifier)
		return repository.Put(msg)
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		repository := bolt.NewMessageRepository(tx, identifier)

		retrievedMessage, err := repository.Get(msg.Id())
		require.NoError(t, err)
		require.Equal(t, retrievedMessage.Raw(), msg.Raw())

		n, err := repository.Count()
		require.NoError(t, err)
		require.Equal(t, 1, n)

		return nil
	})
	require.NoError(t, err)
}

func TestReadMessageRepository_Count(t *testing.T) {
	db := fixtures.Bolt(t)
	identifier := NewRawMessageIdentifierMock()

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	a, err := di.BuildAdaptersForTest(db)
	require.NoError(t, err)

	count, err := a.MessageRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 0, count)

	err = db.Update(func(tx *bbolt.Tx) error {
		repository := bolt.NewMessageRepository(tx, identifier)
		return repository.Put(msg)
	})
	require.NoError(t, err)

	count, err = a.MessageRepository.Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

type RawMessageIdentifierMock struct {
}

func NewRawMessageIdentifierMock() *RawMessageIdentifierMock {
	return &RawMessageIdentifierMock{}
}

func (r RawMessageIdentifierMock) IdentifyRawMessage(raw message.RawMessage) (message.Message, error) {
	return message.NewMessage(
		fixtures.SomeRefMessage(),
		nil,
		message.MustNewSequence(1),
		fixtures.SomeRefAuthor(),
		fixtures.SomeRefFeed(),
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		raw,
	)
}
