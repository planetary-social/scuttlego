package feeds

import (
	"github.com/planetary-social/go-ssb/identity"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
)

type FeedFormat interface {
	Verify(raw message.RawMessage) (message.Message, error)
	Sign(unsigned message.UnsignedMessage, private identity.Private) (message.Message, error)
}
