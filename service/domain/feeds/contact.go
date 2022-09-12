package feeds

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type Contact struct {
	target    refs.Identity
	following bool
	blocking  bool
}

func NewContact(target refs.Identity) (*Contact, error) {
	if target.IsZero() {
		return nil, errors.New("zero value of target")
	}
	return &Contact{
		target: target,
	}, nil
}

func NewContactFromHistory(target refs.Identity, following, blocking bool) (*Contact, error) {
	if target.IsZero() {
		return nil, errors.New("zero value of target")
	}
	return &Contact{
		target:    target,
		following: following,
		blocking:  blocking,
	}, nil
}

func MustNewContactFromHistory(target refs.Identity, following, blocking bool) *Contact {
	v, err := NewContactFromHistory(target, following, blocking)
	if err != nil {
		panic(err)
	}
	return v
}

func (c *Contact) Update(actions content.ContactActions) error {
	if actions.IsZero() {
		return errors.New("zero value of actions")
	}

	for _, action := range actions.List() {
		switch action {
		case content.ContactActionFollow:
			c.following = true
		case content.ContactActionUnfollow:
			c.following = false
		case content.ContactActionBlock:
			c.blocking = true
		case content.ContactActionUnblock:
			c.blocking = false
		default:
			return fmt.Errorf("unknown contact action '%#v'", action)
		}
	}

	return nil
}

func (c *Contact) Target() refs.Identity {
	return c.target
}

func (c *Contact) Following() bool {
	return c.following
}

func (c *Contact) Blocking() bool {
	return c.blocking
}
