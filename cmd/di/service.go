package di

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/network"
	"github.com/planetary-social/go-ssb/scuttlebutt/commands"
)

type Service struct {
	listener network.Listener
	app      commands.Application
}

func NewService(listener network.Listener, app commands.Application) Service {
	return Service{listener: listener, app: app}
}

func (s Service) Run() error {
	errCh := make(chan error)
	runners := 0

	// todo remove when connection manager is available
	//go func() {
	//	<-time.After(5 * time.Second)
	//	_, err := s.ssb.Connect()
	//	if err != nil {
	//		panic(err)
	//	}
	//}()

	runners++
	go func() {
		errCh <- s.listener.ListenAndServe()
	}()

	for i := 0; i < runners; i++ {
		err := <-errCh
		if err != nil {
			return errors.Wrap(err, "something terminated with an error")
		}
	}

	return nil
}
