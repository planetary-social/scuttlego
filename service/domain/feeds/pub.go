package feeds

import (
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type PubToSave struct {
	who     refs.Identity
	id      refs.Message
	content content.Pub
}

func NewPubToSave(who refs.Identity, id refs.Message, content content.Pub) PubToSave {
	return PubToSave{
		who:     who,
		id:      id,
		content: content,
	}
}

func (c PubToSave) Who() refs.Identity {
	return c.who
}

func (c PubToSave) Id() refs.Message {
	return c.id
}

func (c PubToSave) Content() content.Pub {
	return c.content
}
