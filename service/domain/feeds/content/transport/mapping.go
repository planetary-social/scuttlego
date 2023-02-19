package transport

import (
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
)

type MessageContentMappings map[known.MessageContentType]MessageContentMapping

type MessageContentMapping struct {
	Marshal   func(con known.KnownMessageContent) ([]byte, error)
	Unmarshal func(b []byte) (known.KnownMessageContent, error)
}

func DefaultMappings() MessageContentMappings {
	return MessageContentMappings{
		known.Contact{}.Type(): ContactMapping,
		known.Pub{}.Type():     PubMapping,
	}
}
