package transport

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

var aboutMapping = MessageContentMapping{
	Marshal: func(con content.KnownMessageContent) ([]byte, error) {
		return nil, errors.New("not implemented")
	},
	Unmarshal: func(b []byte) (content.KnownMessageContent, error) {
		var t transportAbout

		if err := json.Unmarshal(b, &t); err != nil {
			return nil, errors.Wrap(err, "json unmarshal failed")
		}

		image, err := unmarshalAboutImage(t.Image)
		if err != nil {
			return nil, errors.Wrap(err, "could not create image ref")
		}

		return content.NewAbout(image)
	},
}

func unmarshalAboutImage(j json.RawMessage) (*refs.Blob, error) {
	if len(j) == 0 {
		return nil, nil
	}

	var blobRefString string
	if err := json.Unmarshal(j, &blobRefString); err == nil {
		if blobRefString == "" {
			return nil, nil
		}

		blob, err := refs.NewBlob(blobRefString)
		if err != nil {
			return nil, errors.Wrap(err, "could not create a blob ref")
		}

		return &blob, nil
	}

	mention, err := unmarshalMention(j)
	if err != nil {
		return nil, errors.Wrap(err, "invalid mention")
	}

	blob, err := refs.NewBlob(mention.Link)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a blob ref")
	}

	return &blob, err
}

type transportAbout struct {
	messageContentType // todo this is stupid

	// this may be a plain string with a blob ref in it or a blobLink object
	Image json.RawMessage `json:"image"`
}

type mention struct {
	Link string `json:"link"`
}

func unmarshalMention(j json.RawMessage) (mention, error) {
	var m mention
	if err := json.Unmarshal(j, &m); err != nil {
		return m, errors.Wrap(err, "could not unmarshal blob link")
	}

	if m.Link == "" {
		return m, errors.New("link is empty")
	}

	return m, nil
}
