package di

import (
	"context"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/migrations"
	migrations2 "github.com/planetary-social/scuttlego/service/adapters/migrations"
	"github.com/planetary-social/scuttlego/service/app/commands"
)

//nolint:unused
var migrationsSet = wire.NewSet(
	migrations.NewRunner,
	wire.Bind(new(commands.MigrationsRunner), new(*migrations.Runner)),

	migrations.NewMigrations,

	migrations2.NewBoltProgressStorage,
	wire.Bind(new(migrations.ProgressStorage), new(*migrations2.BoltProgressStorage)),

	wire.Struct(new(commands.Migrations), "*"),
	commands.NewMigrationHandlerImportDataFromGoSSB,

	migrations2.NewGoSSBRepoReader,
	wire.Bind(new(commands.GoSSBRepoReader), new(*migrations2.GoSSBRepoReader)),

	newMigrationsList,
)

func newMigrationsList(
	m commands.Migrations,
	config Config,
) []migrations.Migration {
	return []migrations.Migration{
		{
			Name: "0001_import_data_from_gossb",
			Fn: func(ctx context.Context, state migrations.State) (migrations.State, error) {
				cmd, err := commands.NewImportDataFromGoSSB(config.GoSSBDataDirectory)
				if err != nil {
					return state, errors.Wrap(err, "could not create a command")
				}

				err = m.MigrationImportDataFromGoSSB.Handle(ctx, cmd)
				if err != nil {
					return state, errors.Wrap(err, "could not run a command")
				}

				return state, nil
			},
		},
	}
}
