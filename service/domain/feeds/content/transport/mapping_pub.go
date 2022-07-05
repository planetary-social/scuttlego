package transport

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

var pubMapping = MessageContentMapping{
	Marshal: func(con content.KnownMessageContent) ([]byte, error) {
		msg := con.(content.Pub)

		t := transportPub{
			messageContentType: contentTypeToTransport(msg),
			Address: transportPubAddress{
				Key:  msg.Key().String(),
				Host: msg.Host(),
				Port: msg.Port(),
			},
		}

		return json.Marshal(t)
	},
	Unmarshal: func(b []byte) (content.KnownMessageContent, error) {
		var t transportPub

		if err := json.Unmarshal(b, &t); err != nil {
			return nil, errors.Wrap(err, "json unmarshal failed")
		}

		key, err := refs.NewIdentity(t.Address.Key)
		if err != nil {
			return nil, errors.Wrap(err, "could not create an identity ref")
		}

		return content.NewPub(key, t.Address.Host, t.Address.Port)
	},
}

type transportPub struct {
	messageContentType                     // todo this is stupid
	Address            transportPubAddress `json:"address"`
}

type transportPubAddress struct {
	Key  string `json:"key"`
	Host string `json:"host"`
	Port int    `json:"port"`
}
