package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/adapters/bolt"
	"github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/replication"
	"go.etcd.io/bbolt"
)

//nolint:deadcode,varcheck
var txBoltAdaptersSet = wire.NewSet(
	bolt.NewFeedRepository,
	wire.Bind(new(commands.FeedRepository), new(*bolt.FeedRepository)),

	bolt.NewSocialGraphRepository,
	wire.Bind(new(commands.SocialGraphRepository), new(*bolt.SocialGraphRepository)),

	bolt.NewReceiveLogRepository,
	bolt.NewMessageRepository,
)

//nolint:deadcode,varcheck
var boltAdaptersSet = wire.NewSet(
	bolt.NewBoltFeedMessagesRepository,
	wire.Bind(new(queries.FeedRepository), new(*bolt.BoltFeedMessagesRepository)),

	bolt.NewBoltContactsRepository,
	wire.Bind(new(replication.Storage), new(*bolt.BoltContactsRepository)),

	bolt.NewReceiveLogReadRepository,
	wire.Bind(new(queries.ReceiveLogRepository), new(*bolt.ReceiveLogReadRepository)),

	bolt.NewReadMessageRepository,
	wire.Bind(new(queries.MessageRepository), new(*bolt.ReadMessageRepository)),

	newTxRepositoriesFactory,
)

func newTxRepositoriesFactory(local identity.Public, logger logging.Logger, config Config) bolt.TxRepositoriesFactory {
	return func(tx *bbolt.Tx) (bolt.TxRepositories, error) {
		return BuildTxRepositories(tx, local, logger, config)
	}
}
