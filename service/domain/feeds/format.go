package feeds

import (
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
)

type FeedFormat interface {
	Load(raw message.VerifiedRawMessage) (message.MessageWithoutId, error)
	Verify(raw message.RawMessage) (message.Message, error)
	Sign(unsigned message.UnsignedMessage, private identity.Private) (message.Message, error)
	Peek(raw message.RawMessage) (PeekedMessage, error)
}
