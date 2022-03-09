package main

import (
	"os"
	"runtime/pprof"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/cmd/ssb-test/di"
	"github.com/planetary-social/go-ssb/cmd/ssb-test/storage"
	"github.com/planetary-social/go-ssb/service/domain/identity"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	f, err := os.Create("/tmp/cpu.profile")
	if err != nil {
		return errors.Wrap(err, "could not create cpu profile")
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		return errors.Wrap(err, "could not start cpu profile")
	}

	go func() {
		<-time.After(60 * time.Second)
		pprof.StopCPUProfile()
		panic("profile done")
	}()

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
