package mocks

import "github.com/planetary-social/scuttlego/service/domain/refs"

type WantListRepositoryMock struct {
	wantList map[string]struct{}
}

func NewWantListRepositoryMock() *WantListRepositoryMock {
	return &WantListRepositoryMock{
		wantList: make(map[string]struct{}),
	}
}

func (w WantListRepositoryMock) Contains(id refs.Blob) (bool, error) {
	_, ok := w.wantList[id.String()]
	return ok, nil
}

func (w WantListRepositoryMock) AddBlob(id refs.Blob) {
	w.wantList[id.String()] = struct{}{}
}

func (w WantListRepositoryMock) List() []refs.Blob {
	var result []refs.Blob
	for key := range w.wantList {
		result = append(result, refs.MustNewBlob(key))
	}
	return result
}

func (w WantListRepositoryMock) Delete(id refs.Blob) error {
	delete(w.wantList, id.String())
	return nil
}
