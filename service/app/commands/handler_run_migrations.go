package commands

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/migrations"
)

type MigrationsRunner interface {
	Run(ctx context.Context, migrations migrations.Migrations, progressCallback migrations.ProgressCallback) error
}

type RunMigrations struct {
	progressCallback migrations.ProgressCallback
}

func NewRunMigrations(progressCallback migrations.ProgressCallback) (RunMigrations, error) {
	if progressCallback == nil {
		return RunMigrations{}, errors.New("nil progress callback")
	}
	return RunMigrations{progressCallback: progressCallback}, nil
}

func (cmd RunMigrations) IsZero() bool {
	return cmd == RunMigrations{}
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

func (h RunMigrationsHandler) Run(ctx context.Context, cmd RunMigrations) error {
	if cmd.IsZero() {
		return errors.New("zero value of cmd")
	}

	return h.runner.Run(ctx, h.migrations, cmd.progressCallback)
}
