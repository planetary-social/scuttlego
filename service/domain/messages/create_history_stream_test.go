package messages_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/stretchr/testify/require"
)

func TestTestCreateHistoryStreamSequenceWeFollowProtocolGuide(t *testing.T) {
	seq := fixtures.SomeSequence()

	args, err := messages.NewCreateHistoryStreamArguments(
		fixtures.SomeRefFeed(),
		&seq,
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	b, err := args.MarshalJSON()
	require.NoError(t, err)

	require.Contains(t, string(b), `"sequence"`)
	require.NotContains(t, string(b), `"seq"`)
}

func TestCreateHistoryStreamSequence(t *testing.T) {
	testCases := []struct {
		Name          string
		String        string
		ExpectedError error
	}{
		{
			Name:          "patchwork",
			String:        `[{"id":"@tgzHDm9HEN0k5wFRLFmNPyGZYNF/M5KpkZqCRhgowVE=.ed25519","seq":1537,"live":true,"keys":false}]`,
			ExpectedError: nil,
		},
		{
			Name:          "protocol_guide",
			String:        `[{"id":"@tgzHDm9HEN0k5wFRLFmNPyGZYNF/M5KpkZqCRhgowVE=.ed25519","sequence":1537,"live":true,"keys":false}]`,
			ExpectedError: nil,
		},
		{
			Name:          "both",
			String:        `[{"id":"@tgzHDm9HEN0k5wFRLFmNPyGZYNF/M5KpkZqCRhgowVE=.ed25519","sequence":1537,"seq":1537,"live":true,"keys":false}]`,
			ExpectedError: nil,
		},
		{
			Name:          "both_different",
			String:        `[{"id":"@tgzHDm9HEN0k5wFRLFmNPyGZYNF/M5KpkZqCRhgowVE=.ed25519","sequence":1537,"seq":1234,"live":true,"keys":false}]`,
			ExpectedError: errors.New("inconsistent sequence argument"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			args, err := messages.NewCreateHistoryStreamArgumentsFromBytes([]byte(testCase.String))
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				require.NotNil(t, args.Sequence())
				require.Equal(t, args.Sequence().Int(), 1537)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}

func BenchmarkNewCreateHistoryStreamArgumentsFromBytes(b *testing.B) {
	testCases := []struct {
		Name   string
		String string
	}{
		{
			Name:   "patchwork",
			String: `[{"id":"@tgzHDm9HEN0k5wFRLFmNPyGZYNF/M5KpkZqCRhgowVE=.ed25519","seq":1537,"live":true,"keys":false}]`,
		},
		{
			Name:   "protocol_guide",
			String: `[{"id":"@tgzHDm9HEN0k5wFRLFmNPyGZYNF/M5KpkZqCRhgowVE=.ed25519","sequence":1537,"live":true,"keys":false}]`,
		},
		{
			Name:   "both",
			String: `[{"id":"@tgzHDm9HEN0k5wFRLFmNPyGZYNF/M5KpkZqCRhgowVE=.ed25519","sequence":1537,"seq":1537,"live":true,"keys":false}]`,
		},
	}

	for _, testCase := range testCases {
		b.Run(testCase.Name, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := messages.NewCreateHistoryStreamArgumentsFromBytes([]byte(testCase.String))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkCreateHistoryStreamArguments_MarshalJSON(b *testing.B) {
	allFields, err := messages.NewCreateHistoryStreamArguments(
		fixtures.SomeRefFeed(),
		internal.Ptr(message.MustNewSequence(123)),
		internal.Ptr(1000),
		internal.Ptr(false),
		internal.Ptr(true),
		internal.Ptr(false),
	)
	require.NoError(b, err)

	someFields, err := messages.NewCreateHistoryStreamArguments(
		fixtures.SomeRefFeed(),
		nil,
		internal.Ptr(1000),
		nil,
		nil,
		nil,
	)
	require.NoError(b, err)

	testCases := []struct {
		Name string
		Args messages.CreateHistoryStreamArguments
	}{
		{
			Name: "all_fields",
			Args: allFields,
		},
		{
			Name: "some_fields",
			Args: someFields,
		},
	}

	for _, testCase := range testCases {
		b.Run(testCase.Name, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := testCase.Args.MarshalJSON()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
