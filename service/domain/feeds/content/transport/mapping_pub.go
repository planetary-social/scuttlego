package transport

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

var PubMapping = MessageContentMapping{
	Marshal: func(con known.KnownMessageContent) ([]byte, error) {
		msg := con.(known.Pub)

		t := transportPub{
			MessageContentType: NewMessageContentType(msg),
			Address: transportPubAddress{
				Key:  msg.Key().String(),
				Host: msg.Host(),
				Port: msg.Port(),
			},
		}

		return json.Marshal(t)
	},
	Unmarshal: func(b []byte) (known.KnownMessageContent, error) {
		var t transportPub

		if err := json.Unmarshal(b, &t); err != nil {
			return nil, errors.Wrap(err, "json unmarshal failed")
		}

		key, err := refs.NewIdentity(t.Address.Key)
		if err != nil {
			return nil, errors.Wrap(err, "could not create an identity ref")
		}

		return known.NewPub(key, t.Address.Host, t.Address.Port)
	},
}

type transportPub struct {
	MessageContentType                     // todo this is stupid
	Address            transportPubAddress `json:"address"`
}

type transportPubAddress struct {
	Key  string `json:"key"`
	Host string `json:"host"`
	Port int    `json:"port"`
}
