package di

import (
	"context"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/migrations"
	migrationsadapters "github.com/planetary-social/scuttlego/service/adapters/migrations"
	"github.com/planetary-social/scuttlego/service/app/commands"
)

//nolint:unused
var migrationsSet = wire.NewSet(
	migrations.NewRunner,
	wire.Bind(new(commands.MigrationsRunner), new(*migrations.Runner)),

	migrations.NewMigrations,

	migrationsadapters.NewBoltStorage,
	wire.Bind(new(migrations.Storage), new(*migrationsadapters.BoltStorage)),

	wire.Struct(new(commands.Migrations), "*"),
	commands.NewMigrationHandlerImportDataFromGoSSB,

	migrationsadapters.NewGoSSBRepoReader,
	wire.Bind(new(commands.GoSSBRepoReader), new(*migrationsadapters.GoSSBRepoReader)),

	newMigrationsList,
)

func newMigrationsList(
	m commands.Migrations,
	config Config,
) []migrations.Migration {
	return []migrations.Migration{
		{
			Name: "0001_import_data_from_gossb",
			Fn: func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				cmd, err := commands.NewImportDataFromGoSSB(config.GoSSBDataDirectory)
				if err != nil {
					return errors.Wrap(err, "could not create a command")
				}

				err = m.MigrationImportDataFromGoSSB.Handle(ctx, cmd)
				if err != nil {
					return errors.Wrap(err, "could not run a command")
				}

				return nil
			},
		},
	}
}
