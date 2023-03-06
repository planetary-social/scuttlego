package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/app/queries"
	blobReplication "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var pubSubSet = wire.NewSet(
	requestPubSubSet,
	messagePubSubSet,
	blobDownloadedPubSubSet,
	roomAttendantEventPubSubSet,
	newPeerPubSubSet,
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

var roomAttendantEventPubSubSet = wire.NewSet(
	pubsub.NewRoomAttendantEventPubSub,
	wire.Bind(new(rooms.AttendantEventPublisher), new(*pubsub.RoomAttendantEventPubSub)),
)

var newPeerPubSubSet = wire.NewSet(
	pubsub.NewNewPeerPubSub,
	wire.Bind(new(transport.NewPeerHandler), new(*pubsub.NewPeerPubSub)),
)
