package content

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type About struct {
	image *refs.Blob
}

func NewAbout(image *refs.Blob) (About, error) {
	if image != nil && image.IsZero() {
		return About{}, errors.New("zero value of image")
	}

	return About{image: image}, nil
}

func MustNewAbout(image *refs.Blob) About {
	about, err := NewAbout(image)
	if err != nil {
		panic(err)
	}
	return about
}

func (a About) Type() MessageContentType {
	return "about"
}

func (a About) Blobs() []refs.Blob {
	if a.image != nil {
		return []refs.Blob{*a.image}
	}
	return nil
}
