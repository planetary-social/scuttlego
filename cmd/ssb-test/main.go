package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/cmd/ssb-test/storage"
	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/invites"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

//var (
//myPatchwork        = refs.MustNewIdentity("@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519")
//myPatchworkConnect = commands.Connect{
//	Remote:  myPatchwork.Identity(),
//	Address: network.NewAddress("127.0.0.1:8008"),
//}

//localGoSSB        = refs.MustNewIdentity("@ln1Bdt8lEy4/F/szWlFVAIAIdCBKmzH2MNEVad8BWus=.ed25519")
//localGoSSBConnect = commands.Connect{
//	Remote:  localGoSSB.Identity(),
//	Address: network.NewAddress("127.0.0.1:8008"),
//}

//soapdog = refs.MustNewIdentity("@qv10rF4IsmxRZb7g5ekJ33EakYBpdrmV/vtP1ij5BS4=.ed25519")

//pub         = refs.MustNewIdentity("@CIlwTOK+m6v1hT2zUVOCJvvZq7KE/65ErN6yA2yrURY=.ed25519")
//hubConnect = commands2.Connect{
//	Remote:  pub.Identity(),
//	Address: network2.NewAddress("one.planetary.pub:8008"),
//}
//)

var mainnetPubs = []invites.Invite{
	invites.MustNewInviteFromString("one.planetary.pub:8008:@CIlwTOK+m6v1hT2zUVOCJvvZq7KE/65ErN6yA2yrURY=.ed25519~KVvak/aZeQJQUrn1imLIvwU+EVTkCzGW8TJWTmK8lOk="),
	invites.MustNewInviteFromString("two.planetary.pub:8008:@7jJ7oou5pKKuyKvIlI5tl3ncjEXmZcbm3TvKqQetJIo=.ed25519~8pETEamsgecH32ry4bj7sr7ofXtUbeOCG1qq4C7szHY="),
	invites.MustNewInviteFromString("three.planetary.pub:8008:@LQ8HBiEinU5FiXGaZH9JYFGBGdsB99mepBdh/Smq3VI=.ed25519~tEXSFgAGmGbgb6+lWv5LGNdSWeM5cRjLITnGVrJFfYg="),
	invites.MustNewInviteFromString("four.planetary.pub:8008:@5KDK98cjIQ8bPoBkvp7bCwBXoQMlWpdIbCFyXER8Lbw=.ed25519~e9ZRXEw0RSTE6FX8jOwWV7yfMRDsAZkzlhCRbVMBUEc="),
}

var (
	testnetPub         = invites.MustNewInviteFromString("198.199.90.207:8008:@2xO+nZ1D46RIc6hGKk1fJ4ccynogPNry1S7q18XZQGk=.ed25519~9qgQcC9XngzFLV2A9kIOyVo0q8P+twN6VLKl4DBOgsQ=")
	testnetNetworkKey  boxstream.NetworkKey
	testnetMessageHMAC formats.MessageHMAC
)

func init() {
	keyBytes, err := base64.StdEncoding.DecodeString("AHSrRkNQlQbJP3FyKxvBUI02LI0OdixEl0pYFTUHrMw=")
	if err != nil {
		panic(err)
	}
	testnetNetworkKey = boxstream.MustNewNetworkKey(keyBytes)

	hmacBytes, err := base64.StdEncoding.DecodeString("d8mXyv5OAjxTLnPIAnXUk7TjjALroyfdJn+0RUGHxY4=")
	if err != nil {
		panic(err)
	}
	testnetMessageHMAC = formats.MustNewMessageHMAC(hmacBytes)
}

func mainnetPubsAsPreferredPubs() []domain.Pub {
	var result []domain.Pub
	for _, pub := range mainnetPubs {
		result = append(result, domain.Pub{
			Identity: pub.Remote().Identity(),
			Address:  pub.Address(),
		})
	}
	return result
}

func run() error {
	go captureCPUProfiles()
	go captureHeapProfiles()

	ctx := context.Background()

	if len(os.Args) != 2 {
		return errors.New("invalid arguments")
	}

	config := di.Config{
		DataDirectory: os.Args[1],
		ListenAddress: ":8008",
		Logger:        newLogger(),
		//NetworkKey:    testnetNetworkKey,
		//MessageHMAC:   testnetMessageHMAC,
		PeerManagerConfig: domain.PeerManagerConfig{
			//PreferredPubs: []domain.Pub{
			//	{
			//		Identity: testnetPub.Remote().Identity(),
			//		Address:  testnetPub.Address(),
			//	},
			//},
			PreferredPubs: mainnetPubsAsPreferredPubs(),
		},
	}

	config.SetDefaults()

	local, err := loadOrGenerateIdentity(config)
	if err != nil {
		return errors.Wrap(err, "could not load the identity")
	}

	config.Logger.WithField("identity", local.Public()).Debug("my identity")

	service, err := di.BuildService(ctx, local, config)
	if err != nil {
		return errors.Wrap(err, "could not build a service")
	}

	//go func() {
	//	<-time.After(5 * time.Second)
	//	err := service.App.Commands.Follow.Handle(commands.Follow{Target: testnetPub.Remote()})
	//	fmt.Println("follow", err)
	//}()

	go func() {
		for _, pub := range mainnetPubs {
			err := service.App.Commands.Follow.Handle(commands.Follow{Target: pub.Remote()})
			fmt.Println("follow", pub.Remote(), "err", err)
		}
	}()

	//go func() {
	//	<-time.After(5 * time.Second)
	//	if err := service.App.Commands.Connect.Handle(myPatchworkConnect); err != nil {
	//		fmt.Println("error", err)
	//	}
	//}()

	go func() {
		for {
			<-time.After(1 * time.Second)
			result, err := service.App.Queries.Status.Handle()
			if err != nil {
				panic(err)
			}

			var peers []string
			for _, peer := range result.Peers {
				peers = append(peers, peer.Identity.String())
			}

			config.Logger.
				WithField("feeds", result.NumberOfFeeds).
				WithField("messages", result.NumberOfMessages).
				WithField("peers", strings.Join(peers, ", ")).
				WithField("goroutines", runtime.NumGoroutine()).
				Debug("status")
		}
	}()

	return service.Run(ctx)
}

func newLogger() logging.LogrusLogger {
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	return logging.NewLogrusLogger(log, "main", logging.LevelDebug)
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

func captureCPUProfiles() {
	for {
		if err := captureCPUProfile(); err != nil {
			fmt.Println("failed capturing profile", err)
		}
	}
}

func captureHeapProfiles() {
	for {
		if err := captureHeapProfile(); err != nil {
			fmt.Println("failed capturing profile", err)
		}
		<-time.After(60 * time.Second)
	}
}

func captureCPUProfile() error {
	f, err := os.Create(fmt.Sprintf("/tmp/cpu.profile-%s", nowAsString()))
	if err != nil {
		return errors.Wrap(err, "could not create cpu profile")
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		return errors.Wrap(err, "could not start cpu profile")
	}

	<-time.After(60 * time.Second)
	pprof.StopCPUProfile()
	return nil
}

func captureHeapProfile() error {
	f, err := os.Create(fmt.Sprintf("/tmp/heap.profile-%s", nowAsString()))
	if err != nil {
		return errors.Wrap(err, "could not create cpu profile")
	}
	defer f.Close()

	return pprof.WriteHeapProfile(f)
}

func nowAsString() string {
	return time.Now().Format(time.RFC3339)
}
