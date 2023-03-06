package commands_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	mocks2 "github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestRawMessagePublisher(t *testing.T) {
	feedRepository := mocks2.NewFeedRepositoryMock()

	adapters := commands.Adapters{
		Feed: feedRepository,
	}

	transactionProvider := mocks2.NewMockCommandsTransactionProvider(adapters)

	iden := fixtures.SomePrivateIdentity()
	content := message.MustNewRawContent(fixtures.SomeBytes())

	publisher := commands.NewTransactionRawMessagePublisher(transactionProvider)
	_, err := publisher.Publish(iden, content)
	require.NoError(t, err)

	ref := refs.MustNewIdentityFromPublic(iden.Public())

	require.Equal(t,
		[]mocks2.FeedRepositoryMockUpdateFeedCall{
			{
				Feed: ref.MainFeed(),
			},
		},
		feedRepository.UpdateFeedCalls(),
	)
}
