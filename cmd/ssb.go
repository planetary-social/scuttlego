package main

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/cmd/di"
	"github.com/planetary-social/go-ssb/identity"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	local, err := identity.NewPrivate()
	if err != nil {
		return errors.Wrap(err, "could not create a new identity")
	}

	service, err := di.BuildService(local)
	if err != nil {
		return errors.Wrap(err, "could not build a service")
	}

	return service.Run()
}
