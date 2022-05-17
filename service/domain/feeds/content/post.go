package content

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type Post struct {
	mentions []refs.Blob
}

func NewPost(mentions []refs.Blob) (Post, error) {
	for _, ref := range mentions {
		if ref.IsZero() {
			return Post{}, errors.New("zero value in mentions")
		}
	}

	tmp := make([]refs.Blob, len(mentions))
	copy(tmp, mentions)

	return Post{mentions: tmp}, nil
}

func MustNewPost(mentions []refs.Blob) Post {
	post, err := NewPost(mentions)
	if err != nil {
		panic(err)
	}
	return post
}

func (a Post) Type() MessageContentType {
	return "post"
}

func (a Post) Blobs() []refs.Blob {
	return a.mentions
}
