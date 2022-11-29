package migrations

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/migrations"
	"go.cryptoscope.co/luigi"
	"go.cryptoscope.co/margaret"
	"go.cryptoscope.co/ssb/sbot"
	refs "go.mindeco.de/ssb-refs"
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

	var lastSeq int64 = -1

	src, err := bot.ReceiveLog.Query(
		margaret.SeqWrap(true),
		margaret.Gte(lastSeq),
	)
	if err != nil {
		return state, errors.Wrap(err, "error querying receive log")
	}

	for {
		v, err := src.Next(ctx)
		if err != nil {
			if luigi.IsEOS(err) {
				break
			}
			return state, errors.Wrap(err, "error getting next value")
		}

		if err, ok := v.(error); ok {
			if margaret.IsErrNulled(err) {
				continue
			}
			return state, errors.Wrap(err, "margaret returned an error")
		}

		sw, ok := v.(margaret.SeqWrapper)
		if !ok {
			return state, fmt.Errorf("expected message seq wrapper but got '%T'", v)
		}

		rxLogSeq := sw.Seq()

		msg, ok := sw.Value().(refs.Message)
		if !ok {
			return state, fmt.Errorf("expected message but got '%T'", sw.Value())
		}

		lastSeq = rxLogSeq

		fmt.Println(rxLogSeq, msg.Key().String())
	}

	return nil, nil
}
