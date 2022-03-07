package feeds_test

import (
	"testing"

	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
	"github.com/stretchr/testify/require"
)

func TestAppend(t *testing.T) {
	feed := fixtures.SomeRefFeed()
	author := fixtures.SomeRefAuthor()

	msg1 := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		nil,
		message.MustNewSequence(1),
		author,
		feed,
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		fixtures.SomeRawMessage(),
	)

	prevId := msg1.Id()

	msg2 := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		&prevId,
		message.MustNewSequence(2),
		author,
		feed,
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		fixtures.SomeRawMessage(),
	)

	f, err := feeds.NewFeed(msg1)
	require.NoError(t, err)

	err = f.AppendMessage(msg2)
	require.NoError(t, err)
}
