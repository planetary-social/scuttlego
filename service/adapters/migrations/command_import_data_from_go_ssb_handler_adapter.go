package migrations

import (
	"context"
	"strconv"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/migrations"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/common"
)

const resumeAfterSequenceKey = "resumeAfterSequence"

type CommandImportDataFromGoSSBHandlerAdapter struct {
	directory string
	m         commands.Migrations
}

func NewCommandImportDataFromGoSSBHandlerAdapter(
	directory string,
	m commands.Migrations,
) *CommandImportDataFromGoSSBHandlerAdapter {
	return &CommandImportDataFromGoSSBHandlerAdapter{
		directory: directory,
		m:         m,
	}
}

func (a *CommandImportDataFromGoSSBHandlerAdapter) Fn(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
	resumeAfterSequence, err := a.loadState(state)
	if err != nil {
		return errors.Wrap(err, "error loading state")
	}

	saveResumeAfterSequenceFn := func(sequence common.ReceiveLogSequence) error {
		return a.saveState(sequence, saveStateFunc)
	}

	cmd, err := commands.NewImportDataFromGoSSB(
		a.directory,
		resumeAfterSequence,
		saveResumeAfterSequenceFn,
	)
	if err != nil {
		return errors.Wrap(err, "could not create a command")
	}

	_, err = a.m.MigrationImportDataFromGoSSB.Handle(ctx, cmd)
	if err != nil {
		return errors.Wrap(err, "could not run a command")
	}

	return nil

}

func (a *CommandImportDataFromGoSSBHandlerAdapter) loadState(state migrations.State) (*common.ReceiveLogSequence, error) {
	resumeAfterSequenceString, ok := state[resumeAfterSequenceKey]
	if ok {
		resumeAfterSequenceInt, err := strconv.Atoi(resumeAfterSequenceString)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing resume after sequence string")
		}

		receiveAfterSequence, err := common.NewReceiveLogSequence(resumeAfterSequenceInt)
		if err != nil {
			return nil, errors.Wrap(err, "error creating resume after sequence object")
		}

		return &receiveAfterSequence, nil
	}

	return nil, nil
}

func (a *CommandImportDataFromGoSSBHandlerAdapter) saveState(sequence common.ReceiveLogSequence, saveStateFunc migrations.SaveStateFunc) error {
	return saveStateFunc(migrations.State{
		resumeAfterSequenceKey: strconv.Itoa(sequence.Int()),
	})
}
