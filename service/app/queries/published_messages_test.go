package queries_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/cmd/ssb-test/di"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
)

func TestPublishedMessages(t *testing.T) {
	app, err := di.BuildApplicationForTests()
	require.NoError(t, err)

	testCases := []struct {
		Name string

		StartSeq message.Sequence

		ExpectedError error
	}{
		{
			Name: "valid",

			StartSeq: message.MustNewSequence(10),

			ExpectedError: nil,
		},
		{
			Name: "zero_value_of_sequence",

			StartSeq: message.Sequence{},

			ExpectedError: errors.New("zero value of sequence"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			query := queries.PublishedMessages{
				StartSeq: testCase.StartSeq,
			}

			_, err := app.Queries.PublishedMessages.Handle(query)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}
