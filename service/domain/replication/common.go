package replication

import "github.com/planetary-social/scuttlego/service/domain/feeds/message"

type RawMessageHandler interface {
	Handle(msg message.RawMessage) error
}
