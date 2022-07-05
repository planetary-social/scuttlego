package mux_test

import (
	"context"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
	"github.com/stretchr/testify/require"
)

func TestNewMux(t *testing.T) {
	logger := fixtures.TestLogger(t)

	handlers := []mux.Handler{
		newMockHandler(
			rpc.MustNewProcedure(
				fixtures.SomeProcedureName(),
				rpc.ProcedureTypeAsync,
			),
		),
		newMockHandler(
			rpc.MustNewProcedure(
				fixtures.SomeProcedureName(),
				rpc.ProcedureTypeSource,
			),
		),
	}

	_, err := mux.NewMux(logger, handlers)
	require.NoError(t, err)
}

func TestNewMux_ProcedureNamesMustBeUnique(t *testing.T) {
	logger := fixtures.TestLogger(t)

	name := rpc.MustNewProcedureName([]string{"someProcedure"})

	handlers := []mux.Handler{
		newMockHandler(
			rpc.MustNewProcedure(
				name,
				rpc.ProcedureTypeAsync,
			),
		),
		newMockHandler(
			rpc.MustNewProcedure(
				name,
				rpc.ProcedureTypeSource,
			),
		),
	}

	_, err := mux.NewMux(logger, handlers)
	require.EqualError(t, err, "could not add a handler: handler for method 'someProcedure' was already added")
}

type mockHandler struct {
	procedure rpc.Procedure
}

func newMockHandler(procedure rpc.Procedure) *mockHandler {
	return &mockHandler{procedure: procedure}
}

func (m mockHandler) Procedure() rpc.Procedure {
	return m.procedure
}

func (m mockHandler) Handle(ctx context.Context, rw mux.ResponseWriter, req *rpc.Request) error {
	return nil
}
