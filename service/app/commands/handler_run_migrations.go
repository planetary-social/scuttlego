package commands

import (
	"context"
	"github.com/planetary-social/scuttlego/migrations"
)

type MigrationsRunner interface {
	Run(ctx context.Context, migrations migrations.Migrations) error
}

type RunMigrationsHandler struct {
	runner     MigrationsRunner
	migrations migrations.Migrations
}

func NewRunMigrationsHandler(
	runner MigrationsRunner,
	migrations migrations.Migrations,
) *RunMigrationsHandler {
	return &RunMigrationsHandler{
		runner:     runner,
		migrations: migrations,
	}
}

func (h RunMigrationsHandler) Run(ctx context.Context) error {
	return h.runner.Run(ctx, h.migrations)
}
