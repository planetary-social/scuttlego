package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/cmd/ssb-test/di"
	"github.com/planetary-social/go-ssb/cmd/ssb-test/storage"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var (
//myPatchwork        = refs.MustNewIdentity("@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519")
//myPatchworkConnect = commands.Connect{
//	Remote:  myPatchwork.Identity(),
//	Address: network.NewAddress("127.0.0.1:8008"),
//}

//soapdog = refs.MustNewIdentity("@qv10rF4IsmxRZb7g5ekJ33EakYBpdrmV/vtP1ij5BS4=.ed25519")

//localConnect = commands.Connect{
//	Remote:  myPatchwork.Identity(),
//	Address: network.NewAddress("127.0.0.1:8008"),
//}

//pub         = refs.MustNewIdentity("@CIlwTOK+m6v1hT2zUVOCJvvZq7KE/65ErN6yA2yrURY=.ed25519")
//hubConnect = commands2.Connect{
//	Remote:  pub.Identity(),
//	Address: network2.NewAddress("one.planetary.pub:8008"),
//}
)

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

	if len(os.Args) != 2 {
		return errors.New("invalid arguments")
	}

	config := di.Config{
		LoggingLevel:  logging.LevelTrace,
		DataDirectory: os.Args[1],
		ListenAddress: ":8008",
	}

	local, err := loadOrGenerateIdentity(config)
	if err != nil {
		return errors.Wrap(err, "could not load the identity")
	}

	fmt.Println("my identity is", refs.MustNewIdentityFromPublic(local.Public()).String())

	service, err := di.BuildService(local, config)
	if err != nil {
		return errors.Wrap(err, "could not build a service")
	}

	//go func() {
	//	<-time.After(5 * time.Second)
	//	if err := service.App.Commands.Connect.Handle(myPatchworkConnect); err != nil {
	//		fmt.Println("error", err)
	//	}
	//}()

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
