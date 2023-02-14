package transport

import (
	"encoding/json"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

var contactMapping = MessageContentMapping{
	Marshal: func(con known.KnownMessageContent) ([]byte, error) {
		msg := con.(known.Contact)

		t := transportContact{
			messageContentType: contentTypeToTransport(msg),
			Contact:            msg.Contact().String(),
		}

		err := marshalContactActions(msg.Actions(), &t)
		if err != nil {
			return nil, errors.Wrap(err, "could not marshal contact action")
		}

		return json.Marshal(t)
	},
	Unmarshal: func(b []byte) (known.KnownMessageContent, error) {
		var t transportContact

		if err := json.Unmarshal(b, &t); err != nil {
			return nil, errors.Wrap(err, "json unmarshal failed")
		}

		contact, err := refs.NewIdentity(t.Contact)
		if err != nil {
			return nil, errors.Wrap(err, "could not create a feed ref")
		}

		action, err := unmarshalContactAction(t)
		if err != nil {
			return nil, errors.Wrap(err, "could not unmarshal contact action")
		}

		return known.NewContact(contact, action)
	},
}

func unmarshalContactAction(t transportContact) (known.ContactActions, error) {
	var actions []known.ContactAction

	if v := t.Following; v != nil {
		if *v {
			actions = append(actions, known.ContactActionFollow)
		} else {
			actions = append(actions, known.ContactActionUnfollow)
		}
	}

	if v := t.Blocking; v != nil {
		if *v {
			actions = append(actions, known.ContactActionBlock)
		} else {
			actions = append(actions, known.ContactActionUnblock)
		}
	}

	return known.NewContactActions(actions)
}

func marshalContactActions(actions known.ContactActions, t *transportContact) error {
	for _, action := range actions.List() {
		switch action {
		case known.ContactActionFollow:
			t.Following = internal.Ptr(true)
		case known.ContactActionUnfollow:
			t.Following = internal.Ptr(false)
		case known.ContactActionBlock:
			t.Blocking = internal.Ptr(true)
		case known.ContactActionUnblock:
			t.Blocking = internal.Ptr(false)
		default:
			return fmt.Errorf("unknown contact action '%#v'", action)
		}
	}

	return nil
}

type transportContact struct {
	messageContentType        // todo this is stupid
	Contact            string `json:"contact,omitempty"`
	Following          *bool  `json:"following,omitempty"`
	Blocking           *bool  `json:"blocking,omitempty"`
	// todo pub field
}
