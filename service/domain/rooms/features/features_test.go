package features_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/rooms/features"
	"github.com/stretchr/testify/require"
)

func TestNewFeatures(t *testing.T) {
	testCases := []struct {
		Name          string
		Features      []features.Feature
		ExpectedError error
	}{
		{
			Name:          "zero_value_of_feature",
			Features:      []features.Feature{{}},
			ExpectedError: errors.New("zero value of feature"),
		},
		{
			Name:          "empty_slice",
			Features:      nil,
			ExpectedError: nil,
		},
		{
			Name:          "one_feature",
			Features:      []features.Feature{features.FeatureTunnel},
			ExpectedError: nil,
		},
		{
			Name:          "duplicated_features",
			Features:      []features.Feature{features.FeatureTunnel, features.FeatureTunnel},
			ExpectedError: errors.New("duplicate feature: tunnel"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			features, err := features.NewFeatures(testCase.Features)
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				for _, feature := range testCase.Features {
					require.True(t, features.Contains(feature))
				}
			}
		})
	}
}
