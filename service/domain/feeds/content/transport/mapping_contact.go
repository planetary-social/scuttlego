package transport

import (
	"encoding/json"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

var contactMapping = MessageContentMapping{
	Marshal: func(con content.KnownMessageContent) ([]byte, error) {
		msg := con.(content.Contact)

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
	Unmarshal: func(b []byte) (content.KnownMessageContent, error) {
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

		return content.NewContact(contact, action)
	},
}

func unmarshalContactAction(t transportContact) (content.ContactActions, error) {
	var actions []content.ContactAction

	if v := t.Following; v != nil {
		if *v {
			actions = append(actions, content.ContactActionFollow)
		} else {
			actions = append(actions, content.ContactActionUnfollow)
		}
	}

	if v := t.Blocking; v != nil {
		if *v {
			actions = append(actions, content.ContactActionBlock)
		} else {
			actions = append(actions, content.ContactActionUnblock)
		}
	}

	return content.NewContactActions(actions)
}

func marshalContactActions(actions content.ContactActions, t *transportContact) error {
	for _, action := range actions.List() {
		switch action {
		case content.ContactActionFollow:
			t.Following = internal.Ptr(true)
		case content.ContactActionUnfollow:
			t.Following = internal.Ptr(false)
		case content.ContactActionBlock:
			t.Blocking = internal.Ptr(true)
		case content.ContactActionUnblock:
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
