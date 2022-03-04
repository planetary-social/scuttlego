package transport

import (
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
)

type MessageContentMappings map[message.MessageContentType]MessageContentMapping

type MessageContentMapping struct {
	Marshal   func(con message.MessageContent) ([]byte, error)
	Unmarshal func(b []byte) (message.MessageContent, error)
}

func DefaultMappings() MessageContentMappings {
	return MessageContentMappings{
		content.Contact{}.Type(): contactMapping,
	}
}
