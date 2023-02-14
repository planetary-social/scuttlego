package known

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type Contact struct {
	contact refs.Identity
	actions ContactActions
}

func NewContact(contact refs.Identity, actions ContactActions) (Contact, error) {
	if contact.IsZero() {
		return Contact{}, errors.New("zero value of contact")
	}

	if actions.IsZero() {
		return Contact{}, errors.New("zero value of actions")
	}

	return Contact{
		contact: contact,
		actions: actions,
	}, nil
}

func MustNewContact(contact refs.Identity, actions ContactActions) Contact {
	c, err := NewContact(contact, actions)
	if err != nil {
		panic(err)
	}
	return c
}

func (c Contact) Type() MessageContentType {
	return "contact"
}

func (c Contact) Contact() refs.Identity {
	return c.contact
}

func (c Contact) Actions() ContactActions {
	return c.actions
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

type ContactActions struct {
	actions internal.Set[ContactAction]
}

func NewContactActions(actions []ContactAction) (ContactActions, error) {
	if len(actions) == 0 {
		return ContactActions{}, errors.New("actions can not be empty")
	}

	m := internal.NewSet[ContactAction]()

	for _, action := range actions {
		if action.IsZero() {
			return ContactActions{}, errors.New("zero value of action")
		}

		if m.Contains(action) {
			return ContactActions{}, fmt.Errorf("duplicate action: '%s'", action)
		}

		m.Put(action)
	}

	if m.Contains(ContactActionFollow) && m.Contains(ContactActionUnfollow) {
		return ContactActions{}, errors.New("both follow and unfollow are present")
	}

	if m.Contains(ContactActionBlock) && m.Contains(ContactActionUnblock) {
		return ContactActions{}, errors.New("both block and unblock are present")
	}

	if m.Contains(ContactActionFollow) && m.Contains(ContactActionBlock) {
		return ContactActions{}, errors.New("both follow and block are present")
	}

	return ContactActions{
		actions: m,
	}, nil
}

func MustNewContactActions(actions []ContactAction) ContactActions {
	v, err := NewContactActions(actions)
	if err != nil {
		panic(err)
	}
	return v
}

func (a *ContactActions) List() []ContactAction {
	return a.actions.List()
}

func (a ContactActions) IsZero() bool {
	return a.actions.Len() == 0
}
