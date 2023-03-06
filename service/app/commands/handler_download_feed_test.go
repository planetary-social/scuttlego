package commands_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/stretchr/testify/require"
)

func TestDownloadFeedHandler(t *testing.T) {
	tc, err := di.BuildTestCommands(t)
	require.NoError(t, err)

	feed := fixtures.SomeRefFeed()
	now := time.Now()
	until := now.Add(12 * time.Hour)

	tc.CurrentTimeProvider.CurrentTime = now

	cmd, err := commands.NewDownloadFeed(feed)
	require.NoError(t, err)

	err = tc.DownloadFeed.Handle(cmd)
	require.NoError(t, err)

	require.Equal(t,
		[]mocks.FeedWantListRepositoryMockAddCall{
			{
				Id:    feed,
				Until: until,
			},
		},
		tc.FeedWantListRepository.AddCalls,
	)
}
