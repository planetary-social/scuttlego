package app

import (
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	RedeemInvite *commands.RedeemInviteHandler
	Follow       *commands.FollowHandler
	PublishRaw   *commands.PublishRawHandler

	Connect                  *commands.ConnectHandler
	EstablishNewConnections  *commands.EstablishNewConnectionsHandler
	AcceptNewPeer            *commands.AcceptNewPeerHandler
	ProcessNewLocalDiscovery *commands.ProcessNewLocalDiscoveryHandler

	CreateWants  *commands.CreateWantsHandler
	DownloadBlob *commands.DownloadBlobHandler
	CreateBlob   *commands.CreateBlobHandler
}

type Queries struct {
	CreateHistoryStream  *queries.CreateHistoryStreamHandler
	ReceiveLog           *queries.ReceiveLogHandler
	Status               *queries.StatusHandler
	PublishedMessages    *queries.PublishedMessagesHandler
	GetBlob              *queries.GetBlobHandler
	BlobDownloadedEvents *queries.BlobDownloadedEventsHandler
}
