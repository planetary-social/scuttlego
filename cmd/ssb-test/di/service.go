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
	listener   *networkport.Listener
	pubsub     *pubsubport.PubSub
	advertiser *local.Advertiser
	app        app.Application
}

func NewService(
	listener *networkport.Listener,
	pubsub *pubsubport.PubSub,
	advertiser *local.Advertiser,
	app app.Application,
) Service {
	return Service{
		listener:   listener,
		pubsub:     pubsub,
		advertiser: advertiser,
		app:        app,
	}
}

var (
//myPatchwork = refs.MustNewIdentity("@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519")

//soapdog = refs.MustNewIdentity("@qv10rF4IsmxRZb7g5ekJ33EakYBpdrmV/vtP1ij5BS4=.ed25519")

//localConnect = commands.Connect{
//	Remote:  myPatchwork.Identity(),
//	Address: network.NewAddress("127.0.0.1:8008"),
//}

//pub         = refs.MustNewIdentity("@CIlwTOK+m6v1hT2zUVOCJvvZq7KE/65ErN6yA2yrURY=.ed25519")
//hubConnect = commands2.Connect{
//	Remote:  pub.Identity(),
//	Address: network2.NewAddress("one.planetary.pub:8008"),
//}
)

func (s Service) Run(ctx context.Context) error {
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

	//// todo remove, this is just for testing
	//go func() {
	//	<-time.After(5 * time.Second)

	//	err := s.app.Commands.Connect.Handle(localConnect)
	//	if err != nil {
	//		panic(err)
	//	}
	//}()

	//go func() {
	//	<-time.After(5 * time.Second)

	//	err := s.app.Follow.Handle(commands.Follow{
	//		Target: soapdog,
	//	})
	//	if err != nil {
	//		panic(err)
	//	}
	//}()

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
	}
	return err
}
