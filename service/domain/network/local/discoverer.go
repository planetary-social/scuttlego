package local

import (
	"context"
	"net"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/ssbc/go-netwrap"
	ssbnetwork "github.com/ssbc/go-ssb/network"
)

type IdentityWithAddress struct {
	Remote  identity.Public
	Address network.Address
}

type Discoverer struct {
	discoverer *ssbnetwork.Discoverer
	logger     logging.Logger
}

func NewDiscoverer(local identity.Public, logger logging.Logger) (*Discoverer, error) {
	kp, err := newKeyPair(local)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a new key pair")
	}

	discoverer, err := ssbnetwork.NewDiscoverer(
		kp,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a go-ssb discoverer")
	}

	return &Discoverer{
		discoverer: discoverer,
		logger:     logger.New("discoverer"),
	}, nil

}

func (d *Discoverer) Run(ctx context.Context) <-chan IdentityWithAddress {
	ch := make(chan IdentityWithAddress)

	ssbCh, cancel := d.discoverer.Notify()

	go func() {
		<-ctx.Done()
		cancel()
	}()

	go func() {
		defer close(ch)
		for update := range ssbCh {
			v, err := d.convert(update)
			if err != nil {
				d.logger.WithError(err).Error("conversion failed")
				continue
			}

			select {
			case ch <- v:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

func (d *Discoverer) convert(update net.Addr) (IdentityWithAddress, error) {
	refString := update.(netwrap.Addr).Head().String()
	addrString := update.(netwrap.Addr).Inner().String()

	ref, err := refs.NewIdentity(refString)
	if err != nil {
		return IdentityWithAddress{}, errors.Wrapf(err, "could not create an identity ref from '%s'", refString)
	}

	return IdentityWithAddress{
		Remote:  ref.Identity(),
		Address: network.NewAddress(addrString),
	}, nil
}
