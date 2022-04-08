package app

import (
	"github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/app/queries"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	RedeemInvite             *commands.RedeemInviteHandler
	Follow                   *commands.FollowHandler
	Connect                  *commands.ConnectHandler
	AcceptNewPeer            *commands.AcceptNewPeerHandler
	ProcessNewLocalDiscovery *commands.ProcessNewLocalDiscoveryHandler
}

type Queries struct {
	CreateHistoryStream *queries.CreateHistoryStreamHandler
}
