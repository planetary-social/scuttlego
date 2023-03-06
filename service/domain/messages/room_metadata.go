package messages

import (
	"github.com/boreq/errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/planetary-social/scuttlego/service/domain/rooms/features"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	RoomMetadataProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"room", "metadata"}),
		rpc.ProcedureTypeAsync,
	)
)

func NewRoomMetadata() (*rpc.Request, error) {
	return rpc.NewRequest(
		RoomMetadataProcedure.Name(),
		RoomMetadataProcedure.Typ(),
		[]byte("[]"),
	)
}

type RoomMetadataResponse struct {
	membership bool
	features   features.Features
}

func NewRoomMetadataResponseFromBytes(b []byte) (RoomMetadataResponse, error) {
	var transport roomMetadataTransport
	if err := jsoniter.Unmarshal(b, &transport); err != nil {
		return RoomMetadataResponse{}, errors.Wrap(err, "json unmarshal failed")
	}

	var featuresSlice []features.Feature

	for _, featureString := range transport.Features {
		feature, ok := decodeRoomFeature(featureString)
		if ok {
			featuresSlice = append(featuresSlice, feature)
		}
	}

	features, err := features.NewFeatures(featuresSlice)
	if err != nil {
		return RoomMetadataResponse{}, errors.Wrap(err, "could not create features")
	}

	return RoomMetadataResponse{
		membership: transport.Membership,
		features:   features,
	}, nil
}

func NewRoomMetadataResponse(membership bool, ftrs features.Features) RoomMetadataResponse {
	return RoomMetadataResponse{
		membership: membership,
		features:   ftrs,
	}
}

func (r RoomMetadataResponse) Membership() bool {
	return r.membership
}

func (r RoomMetadataResponse) Features() features.Features {
	return r.features
}

type roomMetadataTransport struct {
	Membership bool     `json:"membership"`
	Features   []string `json:"features"`
}

func decodeRoomFeature(s string) (features.Feature, bool) {
	switch s {
	case "tunnel":
		return features.FeatureTunnel, true
	default:
		return features.Feature{}, false
	}
}
