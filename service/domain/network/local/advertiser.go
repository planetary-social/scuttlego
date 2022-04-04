package local

import (
	"context"
	"net"
	"strconv"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	ssbnetwork "go.cryptoscope.co/ssb/network"
	ssbrefs "go.mindeco.de/ssb-refs"
	"golang.org/x/crypto/ed25519"
)

type Advertiser struct {
	advertiser *ssbnetwork.Advertiser
}

func NewAdvertiser(local identity.Public, address string) (*Advertiser, error) {
	kp, err := newKeyPair(local)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a new key pair")
	}

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, errors.Wrap(err, "could not split host port")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert the port to int")
	}

	ip := net.ParseIP(host)

	advertiser, err := ssbnetwork.NewAdvertiser(
		&net.TCPAddr{
			IP:   ip,
			Port: port,
		},
		kp,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create an advertiser")
	}

	return &Advertiser{
		advertiser: advertiser,
	}, nil

}

func (a *Advertiser) Run(ctx context.Context) error {
	// todo there is no way to see if the advertiser is running correctly or not, the errors will not even be printed
	a.advertiser.Start()
	<-ctx.Done()
	a.advertiser.Stop()
	return nil
}

type keyPair struct {
	id refs.Feed
}

func newKeyPair(local identity.Public) (*keyPair, error) {
	public, err := refs.NewIdentityFromPublic(local)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a new ref")
	}

	return &keyPair{
		id: public.MainFeed(),
	}, nil
}

func (k keyPair) ID() ssbrefs.FeedRef {
	ref, err := ssbrefs.ParseFeedRef(k.id.String())
	if err != nil {
		panic(err) // as programmers like to say: this should never happen
	}
	return ref
}

func (k keyPair) Secret() ed25519.PrivateKey {
	return nil // ssb doesn't actually need the secret
}
