package di

import (
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

	newMigrationsList,
	newMigrationImportDataFromGoSSB,
)

func newMigrationImportDataFromGoSSB(config Config) *migrations2.MigrationImportDataFromGoSSB {
	return migrations2.NewMigrationImportDataFromGoSSB(config.GoSSBDataDirectory)
}

func newMigrationsList(
	importDataFromGoSSB *migrations2.MigrationImportDataFromGoSSB,
) []migrations.Migration {
	return []migrations.Migration{
		{
			Name: "0001_import_data_from_gossb",
			Fn:   importDataFromGoSSB.Run,
		},
	}
}
