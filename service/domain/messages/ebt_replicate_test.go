package messages_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
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

func BenchmarkNewEbtReplicateNotesFromBytes(b *testing.B) {
	for _, numberOfNotes := range []int{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("number_of_notes_%d", numberOfNotes), func(b *testing.B) {
			var notes []messages.EbtReplicateNote
			for i := 0; i < numberOfNotes; i++ {
				note, err := messages.NewEbtReplicateNote(fixtures.SomeRefFeed(), true, true, 0)
				require.NoError(b, err)

				notes = append(notes, note)
			}
			replicateNotes, err := messages.NewEbtReplicateNotes(notes)
			require.NoError(b, err)

			notesJSON, err := replicateNotes.MarshalJSON()
			require.NoError(b, err)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := messages.NewEbtReplicateNotesFromBytes(notesJSON)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkEbtReplicateNotes_MarshalJSON(b *testing.B) {
	for _, numberOfNotes := range []int{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("number_of_notes_%d", numberOfNotes), func(b *testing.B) {
			var notes []messages.EbtReplicateNote
			for i := 0; i < numberOfNotes; i++ {
				note, err := messages.NewEbtReplicateNote(fixtures.SomeRefFeed(), true, true, 0)
				require.NoError(b, err)

				notes = append(notes, note)
			}
			replicateNotes, err := messages.NewEbtReplicateNotes(notes)
			require.NoError(b, err)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := replicateNotes.MarshalJSON()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestEbtReplicateNotes_Empty(t *testing.T) {
	t.Run("not_empty", func(t *testing.T) {
		notes, err := messages.NewEbtReplicateNotes([]messages.EbtReplicateNote{
			messages.MustNewEbtReplicateNote(fixtures.SomeRefFeed(), true, true, 1),
		})
		require.NoError(t, err)
		require.False(t, notes.Empty())
	})

	t.Run("empty", func(t *testing.T) {
		notes, err := messages.NewEbtReplicateNotes(nil)
		require.NoError(t, err)
		require.True(t, notes.Empty())
	})
}

func TestEbtReplicateNotes_MarshalJSON(t *testing.T) {
	feed := refs.MustNewFeed("@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519")

	testCases := []struct {
		Receive       bool
		Replicate     bool
		Seq           int
		ExpectedValue string
	}{
		{
			Receive:       false,
			Replicate:     false,
			Seq:           -1,
			ExpectedValue: `{"@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519":-1}`,
		},
		{
			Receive:       true,
			Replicate:     true,
			Seq:           0,
			ExpectedValue: `{"@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519":0}`,
		},
		{
			Receive:       false,
			Replicate:     true,
			Seq:           0,
			ExpectedValue: `{"@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519":1}`,
		},
		{
			Receive:       true,
			Replicate:     true,
			Seq:           1,
			ExpectedValue: `{"@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519":2}`,
		},
		{
			Receive:       false,
			Replicate:     true,
			Seq:           1,
			ExpectedValue: `{"@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519":3}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.ExpectedValue, func(t *testing.T) {
			note, err := messages.NewEbtReplicateNote(feed, testCase.Receive, testCase.Replicate, testCase.Seq)
			require.NoError(t, err)

			notes, err := messages.NewEbtReplicateNotes([]messages.EbtReplicateNote{note})
			require.NoError(t, err)

			j, err := notes.MarshalJSON()
			require.NoError(t, err)

			require.Equal(t, testCase.ExpectedValue, string(j))
		})
	}
}
