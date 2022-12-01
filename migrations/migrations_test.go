package migrations_test

import (
	"context"
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

	err := r.Runner.Run(ctx, m)
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
		{
			Name: name,
			Fn: func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				passedState = &state
				return nil
			},
		},
	})

	err := r.Runner.Run(ctx, m)
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
		{
			Name: name,
			Fn: func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				passedState = &state
				return nil
			},
		},
	})

	someState := migrations.State{
		fixtures.SomeString(): fixtures.SomeString(),
	}

	r.Storage.returnedState = internal.Ptr(someState)

	err := r.Runner.Run(ctx, m)
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
		{
			Name: name,
			Fn: func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return nil
			},
		},
	})

	err := r.Runner.Run(ctx, m)
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
			{
				Name: name,
				Fn: func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
					return testCase.ReturnedError
				},
			},
		})

		err := r.Runner.Run(ctx, m)
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
		{
			Name: name,
			Fn: func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return nil
			},
		},
	})

	r.Storage.returnedStatus = internal.Ptr(migrations.StatusFinished)

	err := r.Runner.Run(ctx, m)
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
		{
			Name: name,
			Fn: func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return nil
			},
		},
	})

	r.Storage.returnedStatus = internal.Ptr(migrations.StatusFailed)

	err := r.Runner.Run(ctx, m)
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
		{
			Name: name1,
			Fn: func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return nil
			},
		},
		{
			Name: name2,
			Fn: func(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
				return nil
			},
		},
	})

	err := r.Runner.Run(ctx, m)
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

	returnedState  *migrations.State
	returnedStatus *migrations.Status
}

func newStorageMock() *storageMock {
	return &storageMock{}
}

func (s *storageMock) LoadState(name string) (migrations.State, error) {
	s.loadStateCalls = append(s.loadStateCalls, loadStateCall{name: name})
	if s.returnedState == nil {
		return nil, migrations.ErrStateNotFound
	}
	return *s.returnedState, nil
}

func (s *storageMock) SaveState(name string, state migrations.State) error {
	s.saveStateCalls = append(s.saveStateCalls, saveStateCall{name: name, state: state})
	return nil
}

func (s *storageMock) LoadStatus(name string) (migrations.Status, error) {
	s.loadStatusCalls = append(s.loadStatusCalls, loadStatusCall{name: name})
	if s.returnedStatus == nil {
		return migrations.Status{}, migrations.ErrStatusNotFound
	}
	return *s.returnedStatus, nil
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
