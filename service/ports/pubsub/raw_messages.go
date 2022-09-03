package pubsub

import (
	"context"

	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/app"
)

// RawMessageSubscriber receives internal events containing raw messages and passes
// them to the application layer.
type RawMessageSubscriber struct {
	pubsub *pubsub.RawMessagePubSub
	app    app.Application
	log    logging.Logger
}

func NewRawMessageSubscriber(pubsub *pubsub.RawMessagePubSub, app app.Application, log logging.Logger) *RawMessageSubscriber {
	return &RawMessageSubscriber{
		pubsub: pubsub,
		app:    app,
		log:    log.New("raw_message_subscriber"),
	}
}

// Run keeps receiving raw message from the pubsub and passing them to the application
// layer until the context is closed.
func (p *RawMessageSubscriber) Run(ctx context.Context) error {
	rawMessages := p.pubsub.SubscribeToNewRawMessages(ctx)

	for rawMessage := range rawMessages {
		if err := p.app.Commands.RawMessage.Handle(rawMessage); err != nil {
			p.log.WithError(err).Error("failed to handle a raw message")
		}
	}

	return nil
}
