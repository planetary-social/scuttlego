package di

import (
	"context"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/go-ssb/service/app"
	"github.com/planetary-social/go-ssb/service/domain/network/local"
	networkport "github.com/planetary-social/go-ssb/service/ports/network"
	pubsubport "github.com/planetary-social/go-ssb/service/ports/pubsub"
)

type Service struct {
	App app.Application

	listener   *networkport.Listener
	pubsub     *pubsubport.PubSub
	advertiser *local.Advertiser
}

func NewService(
	app app.Application,
	listener *networkport.Listener,
	pubsub *pubsubport.PubSub,
	advertiser *local.Advertiser,
) Service {
	return Service{
		App: app,

		listener:   listener,
		pubsub:     pubsub,
		advertiser: advertiser,
	}
}

func (s Service) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error)
	runners := 0

	runners++
	go func() {
		errCh <- s.listener.ListenAndServe(ctx)
	}()

	runners++
	go func() {
		errCh <- s.pubsub.Run(ctx)
	}()

	runners++
	go func() {
		errCh <- s.advertiser.Run(ctx)
	}()

	var err error
	for i := 0; i < runners; i++ {
		err = multierror.Append(err, errors.Wrap(<-errCh, "error returned by runner"))
		cancel()
	}

	return err
}
