package di

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/network"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/commands"
)

type Service struct {
	listener network.Listener
	app      commands.Application
}

func NewService(listener network.Listener, app commands.Application) Service {
	return Service{listener: listener, app: app}
}

var (
	myPatchwork = refs.MustNewIdentity("@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519")
	pub         = refs.MustNewIdentity("@CIlwTOK+m6v1hT2zUVOCJvvZq7KE/65ErN6yA2yrURY=.ed25519")

	localConnect = commands.Connect{
		Remote:  myPatchwork.Identity(),
		Address: network.NewAddress("127.0.0.1:8008"),
	}

	hubConnect = commands.Connect{
		Remote:  pub.Identity(),
		Address: network.NewAddress("one.planetary.pub:8008"),
	}
)

func (s Service) Run() error {
	errCh := make(chan error)
	runners := 0

	//cmd := commands.RedeemInvite{
	//	Invite: invites.MustNewInviteFromString("one.planetary.pub:8008:@CIlwTOK+m6v1hT2zUVOCJvvZq7KE/65ErN6yA2yrURY=.ed25519~KVvak/aZeQJQUrn1imLIvwU+EVTkCzGW8TJWTmK8lOk="),
	//}

	// todo remove, this is just for testing
	//go func() {
	//	<-time.After(5 * time.Second)

	//	err := s.app.RedeemInvite.Handle(context.Background(), cmd)
	//	if err != nil {
	//		panic(err)
	//	}
	//}()

	// todo remove, this is just for testing
	go func() {
		<-time.After(5 * time.Second)

		err := s.app.Connect.Handle(localConnect)
		if err != nil {
			panic(err)
		}
	}()

	//go func() {
	//	<-time.After(5 * time.Second)

	//	err := s.app.Follow.Handle(commands.Follow{
	//		Target: myPatchwork,
	//	})
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
