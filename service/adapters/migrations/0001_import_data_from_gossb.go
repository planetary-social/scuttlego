package migrations

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/migrations"
	"go.cryptoscope.co/ssb/sbot"
)

type MigrationImportDataFromGoSSB struct {
	directory string
}

func NewMigrationImportDataFromGoSSB(directory string) *MigrationImportDataFromGoSSB {
	return &MigrationImportDataFromGoSSB{directory: directory}
}

func (m MigrationImportDataFromGoSSB) Run(ctx context.Context, state migrations.State) (migrations.State, error) {
	bot, err := sbot.New(
		sbot.WithRepoPath(m.directory),
		//sbot.WithAppKey(),
		//sbot.WithHMACSigning(),
	)
	if err != nil {
		return state, errors.Wrap(err, "error making a bot")
	}

	fmt.Println("seq", bot.ReceiveLog.Seq())

	return nil, nil
}
