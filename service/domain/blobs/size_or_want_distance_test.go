package blobs_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/stretchr/testify/require"
)

func TestSizeOrWantDistance(t *testing.T) {
	testCases := []struct {
		Name                 string
		Value                int64
		ExpectedSize         *blobs.Size
		ExpectedWantDistance *blobs.WantDistance
		ExpectedError        error
	}{
		{
			Name:                 "0_is_invalid",
			Value:                0,
			ExpectedSize:         nil,
			ExpectedWantDistance: nil,
			ExpectedError:        errors.New("0 is neither a distance nor the size"),
		},
		{
			Name:                 "positive_number_is_the_size",
			Value:                1,
			ExpectedSize:         internal.Ptr(blobs.MustNewSize(1)),
			ExpectedWantDistance: nil,
			ExpectedError:        nil,
		},
		{
			Name:                 "negative_number_is_the_want_distance",
			Value:                -1,
			ExpectedSize:         nil,
			ExpectedWantDistance: internal.Ptr(blobs.MustNewWantDistance(1)),
			ExpectedError:        nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			v, err := blobs.NewSizeOrWantDistance(testCase.Value)
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			} else {
				s, ok := v.Size()
				if testCase.ExpectedSize != nil {
					require.True(t, ok)
					require.Equal(t, *testCase.ExpectedSize, s)
				} else {
					require.False(t, ok)
				}

				d, ok := v.WantDistance()
				if testCase.ExpectedWantDistance != nil {
					require.True(t, ok)
					require.Equal(t, *testCase.ExpectedWantDistance, d)
				} else {
					require.False(t, ok)
				}
			}
		})
	}
}
