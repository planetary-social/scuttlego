//go:build wireinject
// +build wireinject

package tests

import (
	"testing"

	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/migrations"
	"github.com/planetary-social/scuttlego/service/app/commands"
)

type CommandIntegrationTest struct {
	DeleteGoSSBRepositoryInOldFormat *commands.MigrationHandlerDeleteGoSSBRepositoryInOldFormat
}

func BuildCommandIntegrationTest(t testing.TB) CommandIntegrationTest {
	wire.Build(
		wire.Struct(new(CommandIntegrationTest), "*"),

		migrations.NewGoSSBRepoReader,
		wire.Bind(new(commands.GoSSBRepoReader), new(*migrations.GoSSBRepoReader)),

		commands.NewMigrationHandlerDeleteGoSSBRepositoryInOldFormat,

		fixtures.TestLogger,
	)

	return CommandIntegrationTest{}
}
