package feeds

import (
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/identity"
)

type FeedFormat interface {
	Verify(raw message.RawMessage) (message.Message, error)
	Sign(unsigned message.UnsignedMessage, private identity.Private) (message.Message, error)
}
