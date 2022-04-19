package transport

import (
	"encoding/json"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

var contactMapping = MessageContentMapping{
	Marshal: func(con content.KnownMessageContent) ([]byte, error) {
		msg := con.(content.Contact)

		t := transportContact{
			messageContentType: contentTypeToTransport(msg),
			Contact:            msg.Contact().String(),
		}

		err := marshalContactAction(msg.Action(), &t)
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

func unmarshalContactAction(t transportContact) (content.ContactAction, error) {
	if t.Following && !t.Blocking {
		return content.ContactActionFollow, nil
	}

	if !t.Following && !t.Blocking {
		return content.ContactActionUnfollow, nil
	}

	if !t.Following && t.Blocking {
		return content.ContactActionBlock, nil
	}

	return content.ContactAction{}, errors.New("invalid contact action")
}

func marshalContactAction(action content.ContactAction, t *transportContact) error {
	switch action {
	case content.ContactActionFollow:
		t.Following = true
		t.Blocking = false
		return nil
	case content.ContactActionUnfollow:
		t.Following = false
		t.Blocking = false
		return nil
	case content.ContactActionBlock:
		t.Following = false
		t.Blocking = true
		return nil
	default:
		return fmt.Errorf("unknown contact action '%#v'", action)
	}
}

type transportContact struct {
	messageContentType        // todo this is stupid
	Contact            string `json:"contact"`
	Following          bool   `json:"following"`
	Blocking           bool   `json:"blocking"`
	// todo pub field
}
