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

	newCommandDeleteGoSsbRepositoryInOldFormatAdapter,
	newCommandImportDataFromGoSSBHandlerAdapter,

	migrationCommandsSet,
)

//nolint:unused
var migrationCommandsSet = wire.NewSet(
	wire.Struct(new(commands.Migrations), "*"),
	commands.NewMigrationHandlerDeleteGoSSBRepositoryInOldFormat,
	commands.NewMigrationHandlerImportDataFromGoSSB,
)

func newCommandDeleteGoSsbRepositoryInOldFormatAdapter(
	config Config,
	m commands.Migrations,
) *migrationsadapters.CommandDeleteGoSsbRepositoryInOldFormatAdapter {
	return migrationsadapters.NewCommandDeleteGoSsbRepositoryInOldFormatAdapter(
		config.GoSSBDataDirectory,
		m,
	)
}

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
	commandDeleteGoSsbRepositoryInOldFormatAdapter *migrationsadapters.CommandDeleteGoSsbRepositoryInOldFormatAdapter,
	commandImportDataFromGoSSBHandlerAdapter *migrationsadapters.CommandImportDataFromGoSSBHandlerAdapter,
) []migrations.Migration {
	return []migrations.Migration{
		migrations.MustNewMigration(
			"delete_gossb_repository_in_old_format",
			commandDeleteGoSsbRepositoryInOldFormatAdapter.Fn,
		),
		migrations.MustNewMigration(
			"import_data_from_gossb",
			commandImportDataFromGoSSBHandlerAdapter.Fn,
		),
	}
}
