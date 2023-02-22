// Package pubsub receives internal events.
package pubsub

import (
	"context"

	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/app/commands"
)

type ProcessRoomAttendantEventHandler interface {
	Handle(ctx context.Context, cmd commands.ProcessRoomAttendantEvent) error
}

type RoomAttendantEventSubscriber struct {
	pubsub  *pubsub.RoomAttendantEventPubSub
	handler ProcessRoomAttendantEventHandler
	logger  logging.Logger
}

func NewRoomAttendantEventSubscriber(
	pubsub *pubsub.RoomAttendantEventPubSub,
	handler ProcessRoomAttendantEventHandler,
	logger logging.Logger,
) *RoomAttendantEventSubscriber {
	return &RoomAttendantEventSubscriber{
		pubsub:  pubsub,
		handler: handler,
		logger:  logger.New("room_attendant_event_subscriber"),
	}
}

func (p *RoomAttendantEventSubscriber) Run(ctx context.Context) error {
	for event := range p.pubsub.SubscribeToAttendantEvents(ctx) {
		cmd, err := commands.NewProcessRoomAttendantEvent(event.Portal, event.Event)
		if err != nil {
			p.logger.WithError(err).Debug("error creating the command")
			continue
		}

		if err := p.handler.Handle(event.Context, cmd); err != nil {
			p.logger.WithError(err).Debug("error handling the command")
			continue
		}
	}

	return nil
}
