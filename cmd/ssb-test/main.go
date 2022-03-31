package main

import (
	"context"
	"os"
	"path"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/cmd/ssb-test/di"
	"github.com/planetary-social/go-ssb/cmd/ssb-test/storage"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/identity"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	//f, err := os.Create("/tmp/cpu.profile")
	//if err != nil {
	//	return errors.Wrap(err, "could not create cpu profile")
	//}
	//defer f.Close()

	//if err := pprof.StartCPUProfile(f); err != nil {
	//	return errors.Wrap(err, "could not start cpu profile")
	//}

	//go func() {
	//	<-time.After(60 * time.Second)
	//	pprof.StopCPUProfile()
	//	panic("profile done")
	//}()

	ctx := context.Background()

	config := di.Config{
		LoggingLevel:  logging.LevelTrace,
		DataDirectory: os.Args[1],
		ListenAddress: ":8008",
	}

	local, err := loadOrGenerateIdentity(config)
	if err != nil {
		return errors.Wrap(err, "could not load the identity")
	}

	service, err := di.BuildService(local, config)
	if err != nil {
		return errors.Wrap(err, "could not build a service")
	}

	return service.Run(ctx)
}

func loadOrGenerateIdentity(config di.Config) (identity.Private, error) {
	filename := path.Join(config.DataDirectory, "identity.json")
	storage := storage.NewIdentityStorage(filename)

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
