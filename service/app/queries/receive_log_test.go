package queries_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/di"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/stretchr/testify/require"
)

func TestReceiveLog(t *testing.T) {
	app, err := di.BuildTestQueries()
	require.NoError(t, err)

	testCases := []struct {
		Name string

		StartSeq int
		Limit    int

		ExpectedError error
	}{
		{
			Name: "valid",

			StartSeq: 0,
			Limit:    1,

			ExpectedError: nil,
		},
		{
			Name: "invalid_limit",

			StartSeq: 0,
			Limit:    0,

			ExpectedError: errors.New("limit must be positive"),
		},
		{
			Name: "invalid_start_seq",

			StartSeq: -1,
			Limit:    1,

			ExpectedError: errors.New("start seq can't be negative"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			query := queries.ReceiveLog{
				StartSeq: testCase.StartSeq,
				Limit:    testCase.Limit,
			}

			_, err := app.Queries.ReceiveLog.Handle(query)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}
