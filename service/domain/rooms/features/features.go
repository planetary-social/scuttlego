package features

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
)

type Features struct {
	features internal.Set[Feature]
}

func NewFeatures(features []Feature) (Features, error) {
	featuresSet := internal.NewSet[Feature]()

	for _, feature := range features {
		if feature.IsZero() {
			return Features{}, errors.New("zero value of feature")
		}

		if featuresSet.Contains(feature) {
			return Features{}, fmt.Errorf("duplicate feature: %+v", feature.s)
		}

		featuresSet.Put(feature)
	}

	return Features{
		features: featuresSet,
	}, nil
}

func (f Features) Contains(feature Feature) bool {
	return f.features.Contains(feature)
}

var (
	FeatureTunnel = Feature{"tunnel"}
)

type Feature struct {
	s string
}

func (f Feature) IsZero() bool {
	return f == Feature{}
}
