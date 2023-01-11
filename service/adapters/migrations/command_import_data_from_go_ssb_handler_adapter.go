package migrations

import (
	"context"
	"strconv"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/migrations"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/common"
)

const resumeFromSequenceKey = "resumeFromSequence"

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
	resumeFromSequence, err := a.loadState(state)
	if err != nil {
		return errors.Wrap(err, "error loading state")
	}

	saveResumeFromSequenceFn := func(sequence common.ReceiveLogSequence) error {
		return a.saveState(sequence, saveStateFunc)
	}

	cmd, err := commands.NewImportDataFromGoSSB(
		a.directory,
		resumeFromSequence,
		saveResumeFromSequenceFn,
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
	resumeFromSequenceString, ok := state[resumeFromSequenceKey]
	if ok {
		resumeFromSequenceInt, err := strconv.Atoi(resumeFromSequenceString)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing resume after sequence string")
		}

		resumeFromSequence, err := common.NewReceiveLogSequence(resumeFromSequenceInt)
		if err != nil {
			return nil, errors.Wrap(err, "error creating resume after sequence object")
		}

		return &resumeFromSequence, nil
	}

	return nil, nil
}

func (a *CommandImportDataFromGoSSBHandlerAdapter) saveState(sequence common.ReceiveLogSequence, saveStateFunc migrations.SaveStateFunc) error {
	return saveStateFunc(migrations.State{
		resumeFromSequenceKey: strconv.Itoa(sequence.Int()),
	})
}
