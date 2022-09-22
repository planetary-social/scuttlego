package messages_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
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

func TestNewEbtReplicateNotesFromBytes(t *testing.T) {
	testCases := []struct {
		Value             int
		ExpectedReceive   bool
		ExpectedReplicate bool
		ExpectedSeq       int
	}{
		{
			Value:             -1,
			ExpectedReceive:   false,
			ExpectedReplicate: false,
			ExpectedSeq:       -1,
		},
		{
			Value:             0,
			ExpectedReceive:   true,
			ExpectedReplicate: true,
			ExpectedSeq:       0,
		},
		{
			Value:             1,
			ExpectedReceive:   false,
			ExpectedReplicate: true,
			ExpectedSeq:       0,
		},
		{
			Value:             2,
			ExpectedReceive:   true,
			ExpectedReplicate: true,
			ExpectedSeq:       1,
		},
		{
			Value:             3,
			ExpectedReceive:   false,
			ExpectedReplicate: true,
			ExpectedSeq:       1,
		},
	}

	for _, testCase := range testCases {
		t.Run(strconv.Itoa(testCase.Value), func(t *testing.T) {
			ref := fixtures.SomeRefFeed()

			b := []byte(fmt.Sprintf(`{"%s":%d}`, ref.String(), testCase.Value))

			notes, err := messages.NewEbtReplicateNotesFromBytes(b)
			require.NoError(t, err)

			t.Log(notes)
			require.Equal(t,
				[]messages.EbtReplicateNote{
					messages.MustNewEbtReplicateNote(
						ref,
						testCase.ExpectedReceive,
						testCase.ExpectedReplicate,
						testCase.ExpectedSeq,
					),
				},
				notes.Notes(),
			)
		})
	}
}
