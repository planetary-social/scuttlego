package migrations

import (
	"github.com/planetary-social/scuttlego/migrations"
)

type BoltProgressStorage struct {
}

func NewBoltProgressStorage() *BoltProgressStorage {
	return &BoltProgressStorage{}
}

func (b BoltProgressStorage) Load(name string) (migrations.Progress, error) {
	return migrations.Progress{}, migrations.ErrProgressNotFound
}

func (b BoltProgressStorage) Save(name string, progress migrations.Progress) error {
	return nil
}
