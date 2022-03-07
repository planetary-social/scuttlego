package main

import (
	"os"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/cmd/di"
	"github.com/planetary-social/go-ssb/cmd/storage"
	"github.com/planetary-social/go-ssb/identity"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	local, err := loadOrGenerateIdentity()
	if err != nil {
		return errors.Wrap(err, "could not load the identity")
	}

	service, err := di.BuildService(local)
	if err != nil {
		return errors.Wrap(err, "could not build a service")
	}

	return service.Run()
}

func loadOrGenerateIdentity() (identity.Private, error) {
	storage := storage.NewIdentityStorage("/tmp/identity.json")

	i, err := storage.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			i, err := identity.NewPrivate()
			if err != nil {
				return identity.Private{}, errors.Wrap(err, "failed to create new identity")
			}

			err = storage.Save(i)
			if err != nil {
				return identity.Private{}, errors.Wrap(err, "failed to save new identity")
			}

			return i, nil
		}
	}

	return i, nil
}
