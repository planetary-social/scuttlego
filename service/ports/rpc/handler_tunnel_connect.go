package rpc

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/rooms/tunnel"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
)

type AcceptTunnelConnectHandler interface {
	Handle(ctx context.Context, cmd commands.AcceptTunnelConnect) error
}

type HandlerTunnelConnect struct {
	handler AcceptTunnelConnectHandler
}

func NewHandlerTunnelConnect(handler AcceptTunnelConnectHandler) *HandlerTunnelConnect {
	return &HandlerTunnelConnect{handler: handler}
}

func (h HandlerTunnelConnect) Procedure() rpc.Procedure {
	return messages.TunnelConnectProcedure
}

func (h HandlerTunnelConnect) Handle(ctx context.Context, s mux.Stream, req *rpc.Request) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	args, err := messages.NewTunnelConnectToTargetArgumentsFromBytes(req.Arguments())
	if err != nil {
		return errors.Wrap(err, "error parsing arguments")
	}

	rwc := tunnel.NewStreamReadWriteCloserAdapter(s, cancel)

	cmd, err := commands.NewAcceptTunnelConnect(args.Origin(), args.Target(), args.Portal(), rwc)
	if err != nil {
		return errors.Wrap(err, "error creating the command")
	}

	err = h.handler.Handle(ctx, cmd)
	if err != nil {
		return errors.Wrap(err, "error executing the command")
	}

	return nil
}
