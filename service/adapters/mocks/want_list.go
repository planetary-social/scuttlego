package mocks

import "github.com/planetary-social/go-ssb/service/domain/refs"

type WantListRepositoryMock struct {
	wantList map[string]struct{}
}

func NewWantListRepositoryMock() *WantListRepositoryMock {
	return &WantListRepositoryMock{
		wantList: make(map[string]struct{}),
	}
}

func (w WantListRepositoryMock) WantListContains(id refs.Blob) (bool, error) {
	_, ok := w.wantList[id.String()]
	return ok, nil
}

func (w WantListRepositoryMock) AddBlob(id refs.Blob) {
	w.wantList[id.String()] = struct{}{}
}
