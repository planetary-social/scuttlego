package migrations

import (
	"github.com/planetary-social/scuttlego/migrations"
)

type BoltStorage struct {
}

func NewBoltStorage() *BoltStorage {
	return &BoltStorage{}
}

func (b BoltStorage) LoadState(name string) (migrations.State, error) {
	return migrations.State{}, migrations.ErrStateNotFound
}

func (b BoltStorage) SaveState(name string, state migrations.State) error {
	return nil
}

func (b BoltStorage) LoadStatus(name string) (migrations.Status, error) {
	return migrations.Status{}, migrations.ErrStatusNotFound
}

func (b BoltStorage) SaveStatus(name string, status migrations.Status) error {
	return nil
}
