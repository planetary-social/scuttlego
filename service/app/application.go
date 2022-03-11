package app

import (
	"github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/app/queries"
)

type Application struct {
	RedeemInvite  *commands.RedeemInviteHandler
	Follow        *commands.FollowHandler
	Connect       *commands.ConnectHandler
	AcceptNewPeer *commands.AcceptNewPeerHandler

	Queries Queries
}

type Queries struct {
	CreateHistoryStream *queries.CreateHistoryStreamHandler
}
