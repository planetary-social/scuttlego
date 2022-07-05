package transport

import (
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
)

type MessageContentMappings map[content.MessageContentType]MessageContentMapping

type MessageContentMapping struct {
	Marshal   func(con content.KnownMessageContent) ([]byte, error)
	Unmarshal func(b []byte) (content.KnownMessageContent, error)
}

func DefaultMappings() MessageContentMappings {
	return MessageContentMappings{
		content.Contact{}.Type(): contactMapping,
		content.Pub{}.Type():     pubMapping,
		content.About{}.Type():   aboutMapping,
		content.Post{}.Type():    postMapping,
	}
}

func contentTypeToTransport(messageContent content.KnownMessageContent) messageContentType {
	return messageContentType{string(messageContent.Type())}
}
