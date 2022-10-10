package transport

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
)

var postMapping = MessageContentMapping{
	Marshal: func(con content.KnownMessageContent) ([]byte, error) {
		return nil, errors.New("not implemented")
	},
	Unmarshal: func(b []byte) (content.KnownMessageContent, error) {
		var t transportPost

		if err := json.Unmarshal(b, &t); err != nil {
			return nil, errors.Wrap(err, "json unmarshal failed")
		}

		blobs, err := unmarshalMentions(t.Mentions)
		if err != nil {
			return nil, errors.Wrap(err, "could not unmarshal mentions")
		}

		return content.NewPost(blobs)
	},
}

type transportPost struct {
	messageContentType // todo this is stupid

	Mentions json.RawMessage `json:"mentions"`
}
