package commands_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestRawMessagePublisher(t *testing.T) {
	feedRepository := mocks.NewFeedRepositoryMock()

	adapters := commands.Adapters{
		Feed: feedRepository,
	}

	transactionProvider := mocks.NewMockTransactionProvider(adapters)

	iden := fixtures.SomePrivateIdentity()
	content := message.MustNewRawMessageContent(fixtures.SomeBytes())

	publisher := commands.NewTransactionRawMessagePublisher(transactionProvider)
	_, err := publisher.Publish(iden, content)
	require.NoError(t, err)

	ref := refs.MustNewIdentityFromPublic(iden.Public())

	require.Equal(t,
		[]mocks.FeedRepositoryMockUpdateFeedCall{
			{
				Feed: ref.MainFeed(),
			},
		},
		feedRepository.UpdateFeedCalls(),
	)
}
