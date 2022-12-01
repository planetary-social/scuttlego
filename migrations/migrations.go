package migrations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/logging"
)

type State map[string]string

// MigrationFunc is executed with the previously saved state. If the migration
// func is executed for the first time then the saved state will be an empty
// map. State is saved by calling the provided function. If a migration function
// returns an error it will be executed again. If a function doesn't return an
// error it should not be executed again.
type MigrationFunc func(ctx context.Context, state State, saveStateFunc SaveStateFunc) error

type SaveStateFunc func(state State) error

type Migration struct {
	Name string
	Fn   MigrationFunc
}

type Migrations struct {
	migrations []Migration
}

func NewMigrations(migrations []Migration) (Migrations, error) {
	names := internal.NewSet[string]()
	for _, migration := range migrations {
		if names.Contains(migration.Name) {
			return Migrations{}, fmt.Errorf("duplicate name '%s'", migration.Name)
		}

		names.Put(migration.Name)
	}

	return Migrations{migrations: migrations}, nil
}

func MustNewMigrations(migrations []Migration) Migrations {
	v, err := NewMigrations(migrations)
	if err != nil {
		panic(err)
	}
	return v
}

func (m Migrations) List() []Migration {
	return m.migrations
}

type Status struct {
	s string
}

var (
	StatusFailed   = Status{"failed"}
	StatusFinished = Status{"finished"}
)

var (
	ErrStateNotFound  = errors.New("state not found")
	ErrStatusNotFound = errors.New("status not found")
)

type Storage interface {
	// LoadState returns ErrStateNotFound if state has not been saved yet.
	LoadState(name string) (State, error)
	SaveState(name string, state State) error

	// LoadStatus returns ErrStatusNotFound if status has not been saved yet.
	LoadStatus(name string) (Status, error)
	SaveStatus(name string, status Status) error
}

type Runner struct {
	storage Storage
	logger  logging.Logger
}

func NewRunner(storage Storage, logger logging.Logger) *Runner {
	return &Runner{storage: storage, logger: logger.New("migrations_runner")}
}

func (r Runner) Run(ctx context.Context, migrations Migrations) error {
	for _, migration := range migrations.List() {
		if err := r.runMigration(ctx, migration); err != nil {
			return errors.Wrapf(err, "error running migration '%s'", migration.Name)
		}
	}
	return nil
}

func (r Runner) runMigration(ctx context.Context, migration Migration) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := r.logger.WithField("migration_name", migration.Name)

	logger.Debug("considering migration")

	shouldRun, err := r.shouldRun(migration)
	if err != nil {
		return errors.Wrap(err, "error checking if migration should be run")
	}

	if !shouldRun {
		logger.Debug("not running this migration")
		return nil
	}

	state, err := r.loadState(migration)
	if err != nil {
		return errors.Wrap(err, "error loading state")
	}

	humanReadableState, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "state json marshal error")
	}

	logger.WithField("state", string(humanReadableState)).Debug("executing migration")

	saveStateFunc := func(state State) error {
		return r.storage.SaveState(migration.Name, state)
	}

	migrationErr := migration.Fn(ctx, state, saveStateFunc)
	saveStateErr := r.storage.SaveStatus(migration.Name, r.statusFromError(migrationErr))

	if migrationErr != nil || saveStateErr != nil {
		var resultErr error
		resultErr = multierror.Append(resultErr, errors.Wrap(migrationErr, "migrations error"))
		resultErr = multierror.Append(resultErr, errors.Wrap(saveStateErr, "error saving state"))
		return resultErr
	}

	return nil
}

func (r Runner) shouldRun(migration Migration) (bool, error) {
	status, err := r.storage.LoadStatus(migration.Name)
	if err != nil {
		if errors.Is(err, ErrStatusNotFound) {
			return true, nil
		}
		return false, errors.Wrap(err, "error loading status")
	}
	return status != StatusFinished, nil
}

func (r Runner) loadState(migration Migration) (State, error) {
	state, err := r.storage.LoadState(migration.Name)
	if err != nil {
		if errors.Is(err, ErrStateNotFound) {
			return make(State), nil
		}
		return nil, errors.Wrap(err, "error loading state")
	}
	return state, nil
}

func (r Runner) statusFromError(err error) Status {
	if err == nil {
		return StatusFinished
	}
	return StatusFailed
}
