package ebt

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/replication"
)

type ContactsStorage interface {
	// GetContacts returns a list of contacts. Contacts are sorted by hops,
	// ascending. Contacts include the local feed.
	GetContacts() ([]replication.Contact, error)
}

type MessageWriter interface {
	WriteMessage(msg message.Message) error
}

type Stream interface {
}

type SessionRunner struct {
}

func NewSessionRunner() *SessionRunner {
	return &SessionRunner{}
}

func (s *SessionRunner) HandleStream(ctx context.Context, stream Stream) error {
	return errors.New("not implemented")
}
