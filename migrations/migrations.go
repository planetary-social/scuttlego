package migrations

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/internal"
)

type State map[string]string

type MigrationFunc func(ctx context.Context, state State) (State, error)

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

func (m Migrations) List() []Migration {
	return m.migrations
}

type Runner struct {
	storage ProgressStorage
}

func NewRunner(storage ProgressStorage) *Runner {
	return &Runner{storage: storage}
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
	progress, err := r.storage.Load(migration.Name)
	if err != nil && !errors.Is(err, ErrProgressNotFound) {
		return errors.Wrap(err, "error loading progress")
	}

	if progress.Status == StatusFinished {
		return nil
	}

	newState, migrationErr := migration.Fn(ctx, r.initializeStateIfEmpty(progress.State))
	saveStateErr := r.saveState(migration, newState, err)

	if migrationErr != nil || saveStateErr != nil {
		var resultErr error
		resultErr = multierror.Append(resultErr, errors.Wrap(migrationErr, "migrations error"))
		resultErr = multierror.Append(resultErr, errors.Wrap(saveStateErr, "error saving state"))
		return resultErr
	}

	return nil
}

func (r Runner) initializeStateIfEmpty(state State) State {
	if state == nil {
		return make(State)
	}
	return state
}

func (r Runner) saveState(migration Migration, state State, err error) error {
	return r.storage.Save(migration.Name, Progress{
		Status: r.statusFromError(err),
		State:  state,
	})
}

func (r Runner) statusFromError(err error) Status {
	if err == nil {
		return StatusFinished
	}
	return StatusFailed
}

type Progress struct {
	Status Status
	State  State
}

type Status struct {
	s string
}

var (
	StatusFailed   = Status{"failed"}
	StatusFinished = Status{"finished"}
)

var ErrProgressNotFound = errors.New("not found")

type ProgressStorage interface {
	// Load returns ErrProgressNotFound if progress has not been saved yet.
	Load(name string) (Progress, error)
	Save(name string, progress Progress) error
}
