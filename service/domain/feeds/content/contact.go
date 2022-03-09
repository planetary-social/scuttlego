package content

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type Contact struct {
	contact refs.Identity
	action  ContactAction
}

func NewContact(contact refs.Identity, action ContactAction) (Contact, error) {
	if contact.IsZero() {
		return Contact{}, errors.New("zero value of contact")
	}

	if action.IsZero() {
		return Contact{}, errors.New("zero value of feed")
	}

	return Contact{
		contact: contact,
		action:  action,
	}, nil
}

func MustNewContact(contact refs.Identity, action ContactAction) Contact {
	c, err := NewContact(contact, action)
	if err != nil {
		panic(err)
	}
	return c
}

func (c Contact) Contact() refs.Identity {
	return c.contact
}

func (c Contact) Action() ContactAction {
	return c.action
}

func (c Contact) Type() message.MessageContentType {
	return "contact"
}

type ContactAction struct {
	s string
}

func (a ContactAction) IsZero() bool {
	return a == ContactAction{}
}

var (
	ContactActionFollow   = ContactAction{"follow"}
	ContactActionUnfollow = ContactAction{"unfollow"}
	ContactActionBlock    = ContactAction{"block"}
	ContactActionUnblock  = ContactAction{"unblock"}
)
