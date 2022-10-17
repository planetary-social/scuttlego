package mux_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewMux(t *testing.T) {
	t.Parallel()

	logger := fixtures.TestLogger(t)

	handlers := []mux.Handler{
		newMockHandler(
			rpc.MustNewProcedure(
				fixtures.SomeProcedureName(),
				rpc.ProcedureTypeAsync,
			),
			nil,
		),
		newMockHandler(
			rpc.MustNewProcedure(
				fixtures.SomeProcedureName(),
				rpc.ProcedureTypeSource,
			),
			nil,
		),
	}

	synchronousHandlers := []mux.SynchronousHandler{
		newMockSynchronousHandler(
			rpc.MustNewProcedure(
				fixtures.SomeProcedureName(),
				rpc.ProcedureTypeAsync,
			),
			nil,
		),
		newMockSynchronousHandler(
			rpc.MustNewProcedure(
				fixtures.SomeProcedureName(),
				rpc.ProcedureTypeSource,
			),
			nil,
		),
	}

	_, err := mux.NewMux(logger, handlers, synchronousHandlers)
	require.NoError(t, err)
}

func TestNewMux_ProcedureNamesMustBeUniqueForHandlers(t *testing.T) {
	t.Parallel()

	logger := fixtures.TestLogger(t)

	name := rpc.MustNewProcedureName([]string{"someProcedure"})

	handlers := []mux.Handler{
		newMockHandler(
			rpc.MustNewProcedure(
				name,
				rpc.ProcedureTypeAsync,
			),
			nil,
		),
		newMockHandler(
			rpc.MustNewProcedure(
				name,
				rpc.ProcedureTypeSource,
			),
			nil,
		),
	}

	_, err := mux.NewMux(logger, handlers, nil)
	require.EqualError(t, err, "could not add a handler: handler is not unique: handler for method 'someProcedure' was already added")
}

func TestNewMux_ProcedureNamesMustBeUniqueForSynchronousHandlers(t *testing.T) {
	t.Parallel()

	logger := fixtures.TestLogger(t)

	name := rpc.MustNewProcedureName([]string{"someProcedure"})

	synchronousHandlers := []mux.SynchronousHandler{
		newMockSynchronousHandler(
			rpc.MustNewProcedure(
				name,
				rpc.ProcedureTypeAsync,
			),
			nil,
		),
		newMockSynchronousHandler(
			rpc.MustNewProcedure(
				name,
				rpc.ProcedureTypeSource,
			),
			nil,
		),
	}

	_, err := mux.NewMux(logger, nil, synchronousHandlers)
	require.EqualError(t, err, "could not add a synchronous handler: handler is not unique: synchronous handler for method 'someProcedure' was already added")
}

func TestNewMux_ProcedureNamesMustBeUniqueForSynchronousHandlersAndHandlers(t *testing.T) {
	t.Parallel()

	logger := fixtures.TestLogger(t)

	name := rpc.MustNewProcedureName([]string{"someProcedure"})

	handlers := []mux.Handler{
		newMockHandler(
			rpc.MustNewProcedure(
				name,
				rpc.ProcedureTypeAsync,
			),
			nil,
		),
	}

	synchronousHandlers := []mux.SynchronousHandler{
		newMockSynchronousHandler(
			rpc.MustNewProcedure(
				name,
				rpc.ProcedureTypeSource,
			),
			nil,
		),
	}

	_, err := mux.NewMux(logger, handlers, synchronousHandlers)
	require.EqualError(t, err, "could not add a synchronous handler: handler is not unique: handler for method 'someProcedure' was already added")
}

func TestNewMux_HandlerDoesNotBlock(t *testing.T) {
	t.Parallel()

	logger := fixtures.TestLogger(t)

	procedure := rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"someProcedure"}),
		rpc.ProcedureTypeAsync,
	)

	delay := 1 * time.Second

	handlers := []mux.Handler{
		newMockHandler(
			procedure,
			func(ctx context.Context, s mux.Stream, req *rpc.Request) error {
				if err := s.WriteMessage(fixtures.SomeBytes()); err != nil {
					t.Fatal(err)
				}
				<-time.After(delay)
				return nil
			},
		),
	}

	m, err := mux.NewMux(logger, handlers, nil)
	require.NoError(t, err)

	ctx := fixtures.TestContext(t)
	s := mocks.NewMockCloserStream()
	req := rpc.MustNewRequest(procedure.Name(), procedure.Typ(), nil)

	start := time.Now()
	m.HandleRequest(ctx, s, req)
	require.Eventually(
		t,
		func() bool {
			return len(s.WrittenMessages()) > 0
		},
		delay, 10*time.Millisecond,
	)
	require.Less(t, time.Since(start), delay)
}

func TestNewMux_SynchronousHandlerBlocks(t *testing.T) {
	t.Parallel()

	logger := fixtures.TestLogger(t)

	procedure := rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"someProcedure"}),
		rpc.ProcedureTypeAsync,
	)

	delay := 1 * time.Second

	synchronousHandlers := []mux.SynchronousHandler{
		newMockSynchronousHandler(
			procedure,
			func(ctx context.Context, s mux.Stream, req *rpc.Request) {
				if err := s.WriteMessage(fixtures.SomeBytes()); err != nil {
					t.Fatal(err)
				}
				<-time.After(delay)
			},
		),
	}

	m, err := mux.NewMux(logger, nil, synchronousHandlers)
	require.NoError(t, err)

	ctx := fixtures.TestContext(t)
	s := mocks.NewMockCloserStream()
	req := rpc.MustNewRequest(procedure.Name(), procedure.Typ(), nil)

	start := time.Now()
	m.HandleRequest(ctx, s, req)
	require.Eventually(
		t,
		func() bool {
			return len(s.WrittenMessages()) > 0
		},
		delay, 10*time.Millisecond,
	)
	require.Greater(t, time.Since(start), delay)
}

type handlerFn func(ctx context.Context, s mux.Stream, req *rpc.Request) error

type mockHandler struct {
	procedure rpc.Procedure
	handlerFn handlerFn
}

func newMockHandler(procedure rpc.Procedure, handlerFn handlerFn) *mockHandler {
	return &mockHandler{procedure: procedure, handlerFn: handlerFn}
}

func (m mockHandler) Procedure() rpc.Procedure {
	return m.procedure
}

func (m mockHandler) Handle(ctx context.Context, s mux.Stream, req *rpc.Request) error {
	if m.handlerFn != nil {
		return m.handlerFn(ctx, s, req)
	}
	return nil
}

type synchronousHandlerFn func(ctx context.Context, s mux.Stream, req *rpc.Request)

type mockSynchronousHandler struct {
	procedure            rpc.Procedure
	synchronousHandlerFn synchronousHandlerFn
}

func newMockSynchronousHandler(procedure rpc.Procedure, synchronousHandlerFn synchronousHandlerFn) *mockSynchronousHandler {
	return &mockSynchronousHandler{procedure: procedure, synchronousHandlerFn: synchronousHandlerFn}
}

func (m mockSynchronousHandler) Procedure() rpc.Procedure {
	return m.procedure
}

func (m mockSynchronousHandler) Handle(ctx context.Context, s mux.CloserStream, req *rpc.Request) {
	if m.synchronousHandlerFn != nil {
		m.synchronousHandlerFn(ctx, s, req)
		return
	}
	if err := s.CloseWithError(nil); err != nil {
		fmt.Println(err)
	}
}
