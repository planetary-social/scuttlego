package replication

import (
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
)

type RawMessageHandler interface {
	Handle(replicatedFrom identity.Public, msg message.RawMessage) error
}

type ContactsStorage interface {
	// GetContacts returns a list of contacts. Contacts are sorted by hops,
	// ascending. Contacts include the local feed.
	GetContacts(peer identity.Public) ([]Contact, error)
}
