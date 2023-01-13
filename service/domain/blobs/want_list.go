package blobs

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type WantList struct {
	wants map[string]WantDistance
}

func NewWantList(l []WantedBlob) (WantList, error) {
	wl := WantList{
		wants: make(map[string]WantDistance),
	}

	for _, wantedBlob := range l {
		if wantedBlob.Id.IsZero() {
			return WantList{}, errors.New("zero value of id")
		}

		if wantedBlob.Distance.IsZero() {
			return WantList{}, errors.New("zero value of distance")
		}

		key := wantedBlob.Id.String()

		if _, ok := wl.wants[key]; ok {
			return WantList{}, errors.New("duplicate entry")
		}

		wl.wants[key] = wantedBlob.Distance
	}

	return wl, nil
}

func (wl WantList) List() []WantedBlob {
	var result []WantedBlob

	for blobRef, distance := range wl.wants {
		result = append(result, WantedBlob{
			Id:       refs.MustNewBlob(blobRef),
			Distance: distance,
		})
	}

	return result
}

func (wl WantList) Len() int {
	return len(wl.wants)
}

type WantedBlob struct {
	Id       refs.Blob
	Distance WantDistance
}
