package migrations

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/migrations"
	"github.com/planetary-social/scuttlego/service/app/commands"
)

type CommandDeleteGoSsbRepositoryInOldFormatAdapter struct {
	directory string
	m         commands.Migrations
}

func NewCommandDeleteGoSsbRepositoryInOldFormatAdapter(
	directory string,
	m commands.Migrations,
) *CommandDeleteGoSsbRepositoryInOldFormatAdapter {
	return &CommandDeleteGoSsbRepositoryInOldFormatAdapter{
		directory: directory,
		m:         m,
	}
}

func (a *CommandDeleteGoSsbRepositoryInOldFormatAdapter) Fn(ctx context.Context, _ migrations.State, _ migrations.SaveStateFunc) error {
	cmd, err := commands.NewDeleteGoSSBRepositoryInOldFormat(a.directory)
	if err != nil {
		return errors.Wrap(err, "could not create a command")
	}

	if err = a.m.MigrationDeleteGoSSBRepositoryInOldFormat.Handle(ctx, cmd); err != nil {
		return errors.Wrap(err, "could not run a command")
	}

	return nil

}
