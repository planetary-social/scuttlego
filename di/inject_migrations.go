package di

import (
	"context"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/migrations"
	migrationsadapters "github.com/planetary-social/scuttlego/service/adapters/migrations"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/common"
)

//nolint:unused
var migrationsSet = wire.NewSet(
	migrations.NewRunner,
	wire.Bind(new(commands.MigrationsRunner), new(*migrations.Runner)),

	migrations.NewMigrations,

	migrationsadapters.NewBoltStorage,
	wire.Bind(new(migrations.Storage), new(*migrationsadapters.BoltStorage)),

	migrationsadapters.NewGoSSBRepoReader,
	wire.Bind(new(commands.GoSSBRepoReader), new(*migrationsadapters.GoSSBRepoReader)),

	newMigrationsList,

	migrationCommandsSet,
)

var migrationCommandsSet = wire.NewSet(
	wire.Struct(new(commands.Migrations), "*"),
	commands.NewMigrationHandlerImportDataFromGoSSB,
)

func newMigrationsList(
	m commands.Migrations,
	config Config,
) []migrations.Migration {
	return []migrations.Migration{
		migrations.MustNewMigration(
			"0001_import_data_from_gossb",
			func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				saveResumeAfterSequenceFn := func(sequence common.ReceiveLogSequence) error {
					return nil
				}

				cmd, err := commands.NewImportDataFromGoSSB(
					config.GoSSBDataDirectory,
					nil,
					saveResumeAfterSequenceFn,
				)
				if err != nil {
					return errors.Wrap(err, "could not create a command")
				}

				err = m.MigrationImportDataFromGoSSB.Handle(ctx, cmd)
				if err != nil {
					return errors.Wrap(err, "could not run a command")
				}

				return nil
			},
		),
	}
}
