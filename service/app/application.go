package app

import "github.com/planetary-social/go-ssb/service/app/commands"

type Application struct {
	RedeemInvite *commands.RedeemInviteHandler
	Follow       *commands.FollowHandler
	Connect      *commands.ConnectHandler
}
