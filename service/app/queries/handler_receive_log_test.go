package queries_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/stretchr/testify/require"
)

func TestReceiveLog(t *testing.T) {
	testCases := []struct {
		Name          string
		Limit         int
		ExpectedError error
	}{
		{
			Name:          "limit_positive",
			Limit:         1,
			ExpectedError: nil,
		},
		{
			Name:          "limit_zero",
			Limit:         0,
			ExpectedError: errors.New("limit must be positive"),
		},
		{
			Name:          "limit_negative",
			Limit:         -1,
			ExpectedError: errors.New("limit must be positive"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			seq := fixtures.SomeReceiveLogSequence()
			q, err := queries.NewReceiveLog(seq, testCase.Limit)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				require.Equal(t, testCase.Limit, q.Limit())
				require.Equal(t, seq, q.StartSeq())
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}
