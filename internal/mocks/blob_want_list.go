package mocks

import "github.com/planetary-social/scuttlego/service/domain/refs"

type BlobWantListRepositoryMock struct {
	wantList map[string]struct{}
}

func NewBlobWantListRepositoryMock() *BlobWantListRepositoryMock {
	return &BlobWantListRepositoryMock{
		wantList: make(map[string]struct{}),
	}
}

func (w BlobWantListRepositoryMock) Contains(id refs.Blob) (bool, error) {
	_, ok := w.wantList[id.String()]
	return ok, nil
}

func (w BlobWantListRepositoryMock) AddBlob(id refs.Blob) {
	w.wantList[id.String()] = struct{}{}
}

func (w BlobWantListRepositoryMock) List() []refs.Blob {
	var result []refs.Blob
	for key := range w.wantList {
		result = append(result, refs.MustNewBlob(key))
	}
	return result
}

func (w BlobWantListRepositoryMock) Delete(id refs.Blob) error {
	delete(w.wantList, id.String())
	return nil
}
