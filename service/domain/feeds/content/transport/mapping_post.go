package transport

import (
	"encoding/json"
	"strings"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/refs"
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

		var blobs []refs.Blob
		for _, rawJSON := range t.Mentions {
			mention, err := unmarshalMention(rawJSON)
			if err != nil {
				return nil, errors.Wrap(err, "could not unmarshal a blob link")
			}

			if !strings.HasPrefix(mention.Link, "&") {
				continue
			}

			blob, err := refs.NewBlob(mention.Link)
			if err != nil {
				return nil, errors.Wrap(err, "could not create a blob ref")
			}

			blobs = append(blobs, blob)
		}

		return content.NewPost(blobs)
	},
}

type transportPost struct {
	messageContentType // todo this is stupid

	Mentions []json.RawMessage `json:"mentions"`
}
