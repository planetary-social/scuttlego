package messages_test

import (
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewEbtReplicateArgumentsFromBytes(t *testing.T) {
	b := []byte(`
[
	{
		"version": 3,
		"format": "classic"
	}
]
`)

	args, err := messages.NewEbtReplicateArgumentsFromBytes(b)
	require.NoError(t, err)
	require.Equal(t, 3, args.Version())
	require.Equal(t, messages.EbtReplicateFormatClassic, args.Format())
}

func TestNewEbtReplicate(t *testing.T) {
	args, err := messages.NewEbtReplicateArguments(3, messages.EbtReplicateFormatClassic)
	require.NoError(t, err)

	req, err := messages.NewEbtReplicate(args)
	require.NoError(t, err)

	require.Equal(t, rpc.MustNewProcedureName([]string{"ebt", "replicate"}), req.Name())
	require.Equal(t, rpc.ProcedureTypeDuplex, req.Type())
	require.Equal(t, `[{"version":3,"format":"classic"}]`, string(req.Arguments()))
}
