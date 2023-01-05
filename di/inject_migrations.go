package di

import (
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

	migrationsadapters.NewBadgerStorage,
	wire.Bind(new(migrations.Storage), new(*migrationsadapters.BadgerStorage)),

	migrationsadapters.NewGoSSBRepoReader,
	wire.Bind(new(commands.GoSSBRepoReader), new(*migrationsadapters.GoSSBRepoReader)),

	newMigrationsList,

	newCommandImportDataFromGoSSBHandlerAdapter,

	migrationCommandsSet,
)

var migrationCommandsSet = wire.NewSet(
	wire.Struct(new(commands.Migrations), "*"),
	commands.NewMigrationHandlerImportDataFromGoSSB,
)

func newCommandImportDataFromGoSSBHandlerAdapter(
	config Config,
	m commands.Migrations,
) *migrationsadapters.CommandImportDataFromGoSSBHandlerAdapter {
	return migrationsadapters.NewCommandImportDataFromGoSSBHandlerAdapter(
		config.GoSSBDataDirectory,
		m,
	)
}

func newMigrationsList(
	commandImportDataFromGoSSBHandlerAdapter *migrationsadapters.CommandImportDataFromGoSSBHandlerAdapter,
) []migrations.Migration {
	return []migrations.Migration{
		migrations.MustNewMigration(
			"0001_import_data_from_gossb",
			commandImportDataFromGoSSBHandlerAdapter.Fn,
		),
	}
}
