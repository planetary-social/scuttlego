package content

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type Pub struct {
	key  refs.Identity
	host string
	port int
}

func NewPub(key refs.Identity, host string, port int) (Pub, error) {
	if key.IsZero() {
		return Pub{}, errors.New("zero value of key")
	}

	return Pub{
		key:  key,
		host: host,
		port: port,
	}, nil
}

func MustNewPub(key refs.Identity, host string, port int) Pub {
	v, err := NewPub(key, host, port)
	if err != nil {
		panic(err)
	}
	return v
}

func (c Pub) Type() MessageContentType {
	return "pub"
}

func (c Pub) Key() refs.Identity {
	return c.key
}

func (c Pub) Host() string {
	return c.host
}

func (c Pub) Port() int {
	return c.port
}
