package adapters_test

import (
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/adapters"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestMessageRepository_GetNoMessage(t *testing.T) {
	bolt := fixtures.Bolt(t)
	identifier := NewRawMessageIdentifierMock()

	err := bolt.View(func(tx *bbolt.Tx) error {
		repository := adapters.NewBoltMessageRepository(tx, identifier)

		_, err := repository.Get(fixtures.SomeRefMessage())
		require.EqualError(t, err, "message not found")

		return nil
	})
	require.NoError(t, err)
}

func TestMessageRepository_Put_Get(t *testing.T) {
	bolt := fixtures.Bolt(t)
	identifier := NewRawMessageIdentifierMock()

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := bolt.Update(func(tx *bbolt.Tx) error {
		repository := adapters.NewBoltMessageRepository(tx, identifier)
		return repository.Put(msg)
	})
	require.NoError(t, err)

	err = bolt.View(func(tx *bbolt.Tx) error {
		repository := adapters.NewBoltMessageRepository(tx, identifier)

		retrievedMessage, err := repository.Get(msg.Id())
		require.NoError(t, err)

		require.Equal(t, retrievedMessage.Raw(), msg.Raw())

		return nil
	})
	require.NoError(t, err)
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
