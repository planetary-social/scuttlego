package feeds

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type Contact struct {
	author    refs.Identity
	target    refs.Identity
	following bool
	blocking  bool
}

func NewContact(author, target refs.Identity) (*Contact, error) {
	if author.IsZero() {
		return nil, errors.New("zero value of author")
	}
	if target.IsZero() {
		return nil, errors.New("zero value of target")
	}
	return &Contact{
		author: author,
		target: target,
	}, nil
}

func NewContactFromHistory(author, target refs.Identity, following, blocking bool) (*Contact, error) {
	c, err := NewContact(author, target)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call the constructor")
	}
	c.following = following
	c.blocking = blocking
	return c, nil
}

func MustNewContactFromHistory(author, contact refs.Identity, following, blocking bool) *Contact {
	v, err := NewContactFromHistory(author, contact, following, blocking)
	if err != nil {
		panic(err)
	}
	return v
}

func (c *Contact) Update(actions known.ContactActions) error {
	if actions.IsZero() {
		return errors.New("zero value of actions")
	}

	for _, action := range actions.List() {
		switch action {
		case known.ContactActionFollow:
			c.following = true
		case known.ContactActionUnfollow:
			c.following = false
		case known.ContactActionBlock:
			c.blocking = true
		case known.ContactActionUnblock:
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
