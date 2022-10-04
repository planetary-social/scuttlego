package rooms

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

func GetMetadata(ctx context.Context, peer transport.Peer) (messages.RoomMetadataResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	req, err := messages.NewRoomMetadata()
	if err != nil {
		return messages.RoomMetadataResponse{}, errors.Wrap(err, "could not create a request")
	}

	stream, err := peer.Conn().PerformRequest(ctx, req)
	if err != nil {
		return messages.RoomMetadataResponse{}, errors.Wrap(err, "could not perform a request")
	}

	for v := range stream.Channel() {
		if err := v.Err; err != nil {
			return messages.RoomMetadataResponse{}, errors.Wrap(err, "received an error")
		}

		metadataResponse, err := messages.NewRoomMetadataResponse(v.Value.Bytes())
		if err != nil {
			return messages.RoomMetadataResponse{}, errors.Wrap(err, "could not parse the response")
		}

		return metadataResponse, nil
	}

	return messages.RoomMetadataResponse{}, errors.New("received no responses")
}
