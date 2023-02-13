package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/app/queries"
	blobReplication "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

//nolint:unused
var pubSubSet = wire.NewSet(
	requestPubSubSet,
	messagePubSubSet,
	blobDownloadedPubSubSet,
	roomAttendantEventPubSubSet,
)

//nolint:unused
var requestPubSubSet = wire.NewSet(
	pubsub.NewRequestPubSub,
	wire.Bind(new(rpc.RequestHandler), new(*pubsub.RequestPubSub)),
)

//nolint:unused
var messagePubSubSet = wire.NewSet(
	pubsub.NewMessagePubSub,
	wire.Bind(new(queries.MessageSubscriber), new(*pubsub.MessagePubSub)),
)

//nolint:unused
var blobDownloadedPubSubSet = wire.NewSet(
	pubsub.NewBlobDownloadedPubSub,
	wire.Bind(new(queries.BlobDownloadedSubscriber), new(*pubsub.BlobDownloadedPubSub)),
	wire.Bind(new(blobReplication.BlobDownloadedPublisher), new(*pubsub.BlobDownloadedPubSub)),
)

//nolint:unused
var roomAttendantEventPubSubSet = wire.NewSet(
	pubsub.NewRoomAttendantEventPubSub,
	wire.Bind(new(rooms.AttendantEventPublisher), new(*pubsub.RoomAttendantEventPubSub)),
)
