package storage

import (
	"encoding/json"
	"os"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/identity"
)

type IdentityStorage struct {
	path string
}

func NewIdentityStorage(path string) IdentityStorage {
	return IdentityStorage{path}
}

func (s IdentityStorage) Load() (identity.Private, error) {
	f, err := os.Open(s.path)
	if err != nil {
		return identity.Private{}, errors.Wrap(err, "failed to open a file")
	}
	defer f.Close()

	var stored storedIdentity
	if err := json.NewDecoder(f).Decode(&stored); err != nil {
		return identity.Private{}, errors.Wrap(err, "failed to decode the identity")
	}

	return identity.NewPrivateFromBytes(stored.Private)
}

// todo save to tmp file first
func (s IdentityStorage) Save(private identity.Private) error {
	f, err := os.Create(s.path)
	if err != nil {
		return errors.Wrap(err, "failed to create a file")
	}
	defer f.Close()

	stored := storedIdentity{
		Private: private.PrivateKey(),
	}

	if err := json.NewEncoder(f).Encode(stored); err != nil {
		return errors.Wrap(err, "failed to encode the identity")
	}

	return nil
}

type storedIdentity struct {
	Private []byte `json:"private"`
}
