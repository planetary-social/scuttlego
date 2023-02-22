package rooms

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/rooms/features"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type MetadataGetter interface {
	GetMetadata(ctx context.Context, peer transport.Peer) (messages.RoomMetadataResponse, error)
}

type AttendantsGetter interface {
	GetAttendants(ctx context.Context, peer transport.Peer) (<-chan RoomAttendantsEvent, error)
}

type AttendantEventPublisher interface {
	PublishAttendantEvent(ctx context.Context, portal transport.Peer, event RoomAttendantsEvent) error
}

type Scanner struct {
	metadataGetter   MetadataGetter
	attendantsGetter AttendantsGetter
	publisher        AttendantEventPublisher
	logger           logging.Logger
}

func NewScanner(
	metadataGetter MetadataGetter,
	attendantsGetter AttendantsGetter,
	publisher AttendantEventPublisher,
	logger logging.Logger,
) *Scanner {
	return &Scanner{
		metadataGetter:   metadataGetter,
		attendantsGetter: attendantsGetter,
		publisher:        publisher,
		logger:           logger.New("room_scanner"),
	}
}

func (s Scanner) Run(ctx context.Context, peer transport.Peer) error {
	ok, err := s.canTunnelConnections(ctx, peer)
	if err != nil {
		return errors.Wrap(err, "error checking if the peer can tunnel connections")
	}

	if !ok {
		return nil
	}

	attendants, err := s.attendantsGetter.GetAttendants(ctx, peer)
	if err != nil {
		return errors.Wrap(err, "failed to get attendants")
	}

	for event := range attendants {
		s.logger.
			WithField("peer", peer).
			WithField("event_type", event.typ).
			WithField("event_ref", event.id.String()).
			Debug("publishing an attendant event")

		if err := s.publisher.PublishAttendantEvent(ctx, peer, event); err != nil {
			return errors.Wrap(err, "error publishing an event")
		}
	}

	return nil
}

func (s Scanner) canTunnelConnections(ctx context.Context, peer transport.Peer) (bool, error) {
	metadata, err := s.metadataGetter.GetMetadata(ctx, peer)
	if err != nil {
		if errors.Is(err, rpc.RemoteError{}) {
			return false, nil // most likely this is not a room at all
		}
		return false, errors.Wrap(err, "error getting metadata")
	}

	return metadata.Features().Contains(features.FeatureTunnel), nil
}
