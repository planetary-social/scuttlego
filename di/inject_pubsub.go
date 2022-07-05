package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/go-ssb/service/adapters/pubsub"
	"github.com/planetary-social/go-ssb/service/app/queries"
	blobReplication "github.com/planetary-social/go-ssb/service/domain/blobs/replication"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

//nolint:deadcode,varcheck
var pubSubSet = wire.NewSet(
	requestPubSubSet,
	messagePubSubSet,
	blobDownloadedPubSubSet,
)

var requestPubSubSet = wire.NewSet(
	pubsub.NewRequestPubSub,
	wire.Bind(new(rpc.RequestHandler), new(*pubsub.RequestPubSub)),
)

var messagePubSubSet = wire.NewSet(
	pubsub.NewMessagePubSub,
	wire.Bind(new(queries.MessageSubscriber), new(*pubsub.MessagePubSub)),
)

var blobDownloadedPubSubSet = wire.NewSet(
	pubsub.NewBlobDownloadedPubSub,
	wire.Bind(new(queries.BlobDownloadedSubscriber), new(*pubsub.BlobDownloadedPubSub)),
	wire.Bind(new(blobReplication.BlobDownloadedPublisher), new(*pubsub.BlobDownloadedPubSub)),
)
