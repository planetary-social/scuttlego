package migrations_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/migrations"
	"github.com/stretchr/testify/require"
)

func TestRunner_MigrationsCanBeEmpty(t *testing.T) {
	r := newTestRunner(t)

	ctx := fixtures.TestContext(t)
	m := migrations.MustNewMigrations(nil)
	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Empty(t, r.Storage.saveStateCalls)
	require.Empty(t, r.Storage.saveStatusCalls)
	require.Empty(t, r.Storage.loadStateCalls)
	require.Empty(t, r.Storage.loadStatusCalls)
}

func TestRunner_MigrationIsExecutedWithEmptyInitializedStateIfNoStateIsSaved(t *testing.T) {
	r := newTestRunner(t)

	name := fixtures.SomeString()

	var passedState *migrations.State

	ctx := fixtures.TestContext(t)
	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(
			name,
			func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				passedState = &state
				return nil
			},
		),
	})
	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Equal(t,
		[]loadStateCall{
			{
				name: name,
			},
		},
		r.Storage.loadStateCalls,
	)
	require.EqualValues(t, internal.Ptr(make(migrations.State)), passedState)
}

func TestRunner_MigrationIsExecutedWithSavedStateIfStateWasSaved(t *testing.T) {
	r := newTestRunner(t)

	name := fixtures.SomeString()

	var passedState *migrations.State

	ctx := fixtures.TestContext(t)
	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(
			name,
			func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				passedState = &state
				return nil
			},
		),
	})

	someState := migrations.State{
		fixtures.SomeString(): fixtures.SomeString(),
	}

	r.Storage.MockState(name, someState)
	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Equal(t,
		[]loadStateCall{
			{
				name: name,
			},
		},
		r.Storage.loadStateCalls,
	)
	require.EqualValues(t, internal.Ptr(someState), passedState)
}

func TestRunner_MigrationIsExecutedIfNoStatusIsSaved(t *testing.T) {
	r := newTestRunner(t)

	name := fixtures.SomeString()

	ctx := fixtures.TestContext(t)
	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(
			name,
			func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return nil
			},
		),
	})
	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Equal(t,
		[]loadStateCall{
			{
				name: name,
			},
		},
		r.Storage.loadStateCalls,
	)
	require.Equal(t,
		[]loadStatusCall{
			{
				name: name,
			},
		},
		r.Storage.loadStatusCalls,
	)
}

func TestRunner_StatusIsSavedBasedOnReturnedErrors(t *testing.T) {
	testCases := []struct {
		Name           string
		ReturnedError  error
		ExpectedStatus migrations.Status
	}{
		{
			Name:           "no_error",
			ReturnedError:  nil,
			ExpectedStatus: migrations.StatusFinished,
		},
		{
			Name:           "error",
			ReturnedError:  fixtures.SomeError(),
			ExpectedStatus: migrations.StatusFailed,
		},
	}

	for _, testCase := range testCases {
		r := newTestRunner(t)
		name := fixtures.SomeString()

		ctx := fixtures.TestContext(t)
		m := migrations.MustNewMigrations([]migrations.Migration{
			migrations.MustNewMigration(
				name,
				func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
					return testCase.ReturnedError
				},
			),
		})
		callback := newProgressCallbackMock()

		err := r.Runner.Run(ctx, m, callback)
		if testCase.ReturnedError == nil {
			require.NoError(t, err)
		} else {
			require.ErrorIs(t, err, testCase.ReturnedError)
		}

		require.Equal(t,
			[]saveStatusCall{
				{
					name:   name,
					status: testCase.ExpectedStatus,
				},
			},
			r.Storage.saveStatusCalls,
		)
	}
}

func TestRunner_MigrationIsNotExecutedIfItPreviouslySucceeded(t *testing.T) {
	r := newTestRunner(t)

	name := fixtures.SomeString()

	ctx := fixtures.TestContext(t)
	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(
			name,
			func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return nil
			},
		),
	})

	r.Storage.MockStatus(name, migrations.StatusFinished)
	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Empty(t, r.Storage.loadStateCalls)
	require.Equal(t,
		[]loadStatusCall{
			{
				name: name,
			},
		},
		r.Storage.loadStatusCalls,
	)
}

func TestRunner_MigrationIsExecutedIfItPreviouslyFailed(t *testing.T) {
	r := newTestRunner(t)

	name := fixtures.SomeString()

	ctx := fixtures.TestContext(t)
	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(
			name,
			func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return nil
			},
		),
	})

	r.Storage.MockStatus(name, migrations.StatusFailed)
	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Equal(t,
		[]loadStateCall{
			{
				name: name,
			},
		},
		r.Storage.loadStateCalls,
	)
	require.Equal(t,
		[]loadStatusCall{
			{
				name: name,
			},
		},
		r.Storage.loadStatusCalls,
	)
}

func TestRunner_MigrationsAreConsideredInOrder(t *testing.T) {
	r := newTestRunner(t)

	name1 := fixtures.SomeString()
	name2 := fixtures.SomeString()

	ctx := fixtures.TestContext(t)
	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(
			name1,
			func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return nil
			},
		),
		migrations.MustNewMigration(
			name2,
			func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return nil
			},
		),
	})
	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Equal(t,
		[]loadStateCall{
			{
				name: name1,
			},
			{
				name: name2,
			},
		},
		r.Storage.loadStateCalls,
	)
	require.Equal(t,
		[]loadStatusCall{
			{
				name: name1,
			},
			{
				name: name2,
			},
		},
		r.Storage.loadStatusCalls,
	)
}

func TestRunner_MigrationsCanSaveState(t *testing.T) {
	r := newTestRunner(t)

	name := fixtures.SomeString()
	someState := migrations.State{
		fixtures.SomeString(): fixtures.SomeString(),
	}
	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(
			name,
			func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return saveStateFunc(someState)
			},
		),
	})

	ctx := fixtures.TestContext(t)
	callback := newProgressCallbackMock()
	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Equal(t,
		[]saveStateCall{
			{
				name:  name,
				state: someState,
			},
		},
		r.Storage.saveStateCalls,
	)
}

func TestRunner_IfMigrationsAreEmptyOnlyOnDoneProgressCallbackIsCalled(t *testing.T) {
	r := newTestRunner(t)

	ctx := fixtures.TestContext(t)
	m := migrations.MustNewMigrations(nil)
	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Empty(t, callback.OnRunningCalls)
	require.Empty(t, callback.OnErrorCalls)
	require.Equal(t,
		[]progressCallbackMockOnDoneCall{
			{
				MigrationsCount: 0,
			},
		},
		callback.OnDoneCalls,
	)
}

func TestRunner_OnRunningProgressCallbackIsCalledWhenMigrationsAreRun(t *testing.T) {
	r := newTestRunner(t)

	ctx := fixtures.TestContext(t)

	noop := func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
		return nil
	}

	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(fixtures.SomeString(), noop),
		migrations.MustNewMigration(fixtures.SomeString(), noop),
	})

	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Equal(t,
		[]progressCallbackMockOnRunningCall{
			{
				MigrationIndex:  0,
				MigrationsCount: 2,
			},
			{
				MigrationIndex:  1,
				MigrationsCount: 2,
			},
		},
		callback.OnRunningCalls,
	)
	require.Empty(t, callback.OnErrorCalls)
	require.Equal(t,
		[]progressCallbackMockOnDoneCall{
			{
				MigrationsCount: 2,
			},
		},
		callback.OnDoneCalls,
	)
}

func TestRunner_OnRunningProgressCallbackIsOnlyCalledIfMigrationNeedsToBeRun(t *testing.T) {
	r := newTestRunner(t)

	ctx := fixtures.TestContext(t)

	name1 := fixtures.SomeString()
	name2 := fixtures.SomeString()
	name3 := fixtures.SomeString()

	noop := func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
		return nil
	}

	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(name1, noop),
		migrations.MustNewMigration(name2, noop),
		migrations.MustNewMigration(name3, noop),
	})

	r.Storage.MockStatus(name1, migrations.StatusFinished)

	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.NoError(t, err)

	require.Equal(t,
		[]progressCallbackMockOnRunningCall{
			{
				MigrationIndex:  1,
				MigrationsCount: 3,
			},
			{
				MigrationIndex:  2,
				MigrationsCount: 3,
			},
		},
		callback.OnRunningCalls,
	)
	require.Empty(t, callback.OnErrorCalls)
	require.Equal(t,
		[]progressCallbackMockOnDoneCall{
			{
				MigrationsCount: 3,
			},
		},
		callback.OnDoneCalls,
	)
}

func TestRunner_BothOnProgressAndOnErrorProgressCallbacksAreCalledIfMigrationFails(t *testing.T) {
	r := newTestRunner(t)

	ctx := fixtures.TestContext(t)

	name1 := fixtures.SomeString()
	name2 := fixtures.SomeString()
	someError := fixtures.SomeError()

	noop := func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
		return nil
	}

	returnsError := func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
		return someError
	}

	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(name1, noop),
		migrations.MustNewMigration(name2, returnsError),
	})

	callback := newProgressCallbackMock()

	err := r.Runner.Run(ctx, m, callback)
	require.Error(t, err)

	require.Equal(t,
		[]progressCallbackMockOnRunningCall{
			{
				MigrationIndex:  0,
				MigrationsCount: 2,
			},
			{
				MigrationIndex:  1,
				MigrationsCount: 2,
			},
		},
		callback.OnRunningCalls,
	)
	require.Len(t, callback.OnErrorCalls, 1)
	require.Equal(t, 1, callback.OnErrorCalls[0].MigrationIndex)
	require.Equal(t, 2, callback.OnErrorCalls[0].MigrationsCount)
	require.EqualError(t,
		callback.OnErrorCalls[0].Err,
		fmt.Sprintf(
			"error running migration '%s': 1 error occurred:\n\t* migration function returned an error: %s\n\n",
			name2,
			someError.Error(),
		),
	)
	require.Empty(t, callback.OnDoneCalls)
}

func TestRunner_OnlyOnErrorProgressCallbackIsCalledIfStatusLoadingFails(t *testing.T) {
	r := newTestRunner(t)

	ctx := fixtures.TestContext(t)

	name1 := fixtures.SomeString()
	name2 := fixtures.SomeString()
	someError := fixtures.SomeError()

	noop := func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
		return nil
	}

	m := migrations.MustNewMigrations([]migrations.Migration{
		migrations.MustNewMigration(name1, noop),
		migrations.MustNewMigration(name2, noop),
	})

	callback := newProgressCallbackMock()

	r.Storage.MockLoadStatusError(name2, someError)

	err := r.Runner.Run(ctx, m, callback)
	require.Error(t, err)

	require.Equal(t,
		[]progressCallbackMockOnRunningCall{
			{
				MigrationIndex:  0,
				MigrationsCount: 2,
			},
		},
		callback.OnRunningCalls,
	)
	require.Len(t, callback.OnErrorCalls, 1)
	require.Equal(t, 1, callback.OnErrorCalls[0].MigrationIndex)
	require.Equal(t, 2, callback.OnErrorCalls[0].MigrationsCount)
	require.EqualError(t,
		callback.OnErrorCalls[0].Err,
		fmt.Sprintf(
			"error running migration '%s': error checking if migration should be run: error loading status: %s",
			name2,
			someError.Error(),
		),
	)
	require.Empty(t, callback.OnDoneCalls)
}

type testRunner struct {
	Runner  *migrations.Runner
	Storage *storageMock
}

func newTestRunner(t *testing.T) testRunner {
	logger := fixtures.TestLogger(t)
	storage := newStorageMock()
	runner := migrations.NewRunner(storage, logger)

	return testRunner{
		Runner:  runner,
		Storage: storage,
	}
}

type storageMock struct {
	loadStateCalls  []loadStateCall
	loadStatusCalls []loadStatusCall
	saveStateCalls  []saveStateCall
	saveStatusCalls []saveStatusCall

	returnedState    map[string]migrations.State
	returnedStatus   map[string]migrations.Status
	loadStatusErrors map[string]error
}

func newStorageMock() *storageMock {
	return &storageMock{
		returnedStatus:   make(map[string]migrations.Status),
		returnedState:    make(map[string]migrations.State),
		loadStatusErrors: make(map[string]error),
	}
}

func (s *storageMock) MockState(name string, state migrations.State) {
	s.returnedState[name] = state
}

func (s *storageMock) MockStatus(name string, status migrations.Status) {
	s.returnedStatus[name] = status
}

func (s *storageMock) MockLoadStatusError(name string, err error) {
	s.loadStatusErrors[name] = err
}

func (s *storageMock) LoadState(name string) (migrations.State, error) {
	s.loadStateCalls = append(s.loadStateCalls, loadStateCall{name: name})
	state, ok := s.returnedState[name]
	if !ok {
		return nil, migrations.ErrStateNotFound
	}
	return state, nil
}

func (s *storageMock) SaveState(name string, state migrations.State) error {
	s.saveStateCalls = append(s.saveStateCalls, saveStateCall{name: name, state: state})
	return nil
}

func (s *storageMock) LoadStatus(name string) (migrations.Status, error) {
	s.loadStatusCalls = append(s.loadStatusCalls, loadStatusCall{name: name})

	if err := s.loadStatusErrors[name]; err != nil {
		return migrations.Status{}, err
	}

	status, ok := s.returnedStatus[name]
	if !ok {
		return migrations.Status{}, migrations.ErrStatusNotFound
	}
	return status, nil
}

func (s *storageMock) SaveStatus(name string, status migrations.Status) error {
	s.saveStatusCalls = append(s.saveStatusCalls, saveStatusCall{name: name, status: status})
	return nil
}

type loadStateCall struct {
	name string
}

type saveStateCall struct {
	name  string
	state migrations.State
}

type loadStatusCall struct {
	name string
}

type saveStatusCall struct {
	name   string
	status migrations.Status
}

func TestNewMigrations_DuplicateNamesAreNotAllowed(t *testing.T) {
	name := "some name"

	_, err := migrations.NewMigrations(
		[]migrations.Migration{
			migrations.MustNewMigration(
				name,
				func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
					return nil
				},
			),
			migrations.MustNewMigration(
				name,
				func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
					return nil
				},
			),
		},
	)
	require.EqualError(t, err, "duplicate name 'some name'")
}

func TestNewMigrations_ZeroValuesOfMigrationsAreNotAllowed(t *testing.T) {
	_, err := migrations.NewMigrations(
		[]migrations.Migration{
			{},
		},
	)
	require.EqualError(t, err, "zero value of migration")
}

type progressCallbackMockOnRunningCall struct {
	MigrationIndex  int
	MigrationsCount int
}

type progressCallbackMockOnErrorCall struct {
	MigrationIndex  int
	MigrationsCount int
	Err             error
}

type progressCallbackMockOnDoneCall struct {
	MigrationsCount int
}

type progressCallbackMock struct {
	OnRunningCalls []progressCallbackMockOnRunningCall
	OnErrorCalls   []progressCallbackMockOnErrorCall
	OnDoneCalls    []progressCallbackMockOnDoneCall
}

func newProgressCallbackMock() *progressCallbackMock {
	return &progressCallbackMock{}
}

func (p *progressCallbackMock) OnRunning(migrationIndex int, migrationsCount int) {
	p.OnRunningCalls = append(p.OnRunningCalls, progressCallbackMockOnRunningCall{
		MigrationIndex:  migrationIndex,
		MigrationsCount: migrationsCount,
	})
}

func (p *progressCallbackMock) OnError(migrationIndex int, migrationsCount int, err error) {
	p.OnErrorCalls = append(p.OnErrorCalls, progressCallbackMockOnErrorCall{
		MigrationIndex:  migrationIndex,
		MigrationsCount: migrationsCount,
		Err:             err,
	})
}

func (p *progressCallbackMock) OnDone(migrationsCount int) {
	p.OnDoneCalls = append(p.OnDoneCalls, progressCallbackMockOnDoneCall{
		MigrationsCount: migrationsCount,
	})
}
