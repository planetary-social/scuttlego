package local

import (
	"fmt"
	"net"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/identity"
)

type BroadcasterOptions struct {
	Interval       time.Duration
	BroadcastAddrs []net.IP
	BroadcastPort  int
}

func (o *BroadcasterOptions) setDefault() {
	if o.Interval == 0 {
		o.Interval = time.Second
	}
}

type Broadcaster struct {
	local   identity.Public
	options BroadcasterOptions
}

func NewBroadcaster(local identity.Public, options BroadcasterOptions) (Broadcaster, error) {
	options.setDefault()

	return Broadcaster{
		local:   local,
		options: options,
	}, nil
}

func Run() {

}

func determineBroadcastAddresses() error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return errors.Wrap(err, "could not list interface addresses")
	}

	var broadcastAddrs []net.IP
	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			if v.IP.IsLoopback() {
				continue
			}

			broadcastAddr, err := calculateBroadcast(v.IP, v.Mask)
			if err != nil {
				if errors.Is(err, errNetworkHasNoBroadcast) {
					continue
				}

				return errors.Wrapf(err, "could not calculate a broadcast address for ip '%s'", v.String())
			}

			broadcastAddrs = append(broadcastAddrs, broadcastAddr)
		}
	}

	return nil
}

var errNetworkHasNoBroadcast = errors.New("network has no broadcast")

func calculateBroadcast(addr net.IP, mask net.IPMask) (net.IP, error) {
	var convertedAddress []byte

	if v := addr.To4(); len(v) == net.IPv4len {
		convertedAddress = v
	} else {
		if v := addr.To16(); len(v) == net.IPv6len {
			convertedAddress = v
		} else {
			return nil, fmt.Errorf("unknown addr '%s'", addr.String())
		}
	}

	if len(convertedAddress) != len(mask) {
		return nil, errors.New("mask length doesn't match addr length")
	}

	if ones, bits := mask.Size(); ones == bits {
		return nil, errNetworkHasNoBroadcast
	}

	broadcast := make(net.IP, len(convertedAddress))
	for i := range convertedAddress {
		broadcast[i] = convertedAddress[i] | ^mask[i]
	}

	return broadcast, nil
}
