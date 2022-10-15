package rpc

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
)

type EbtReplicateCommandHandler interface {
	Handle(ctx context.Context, cmd commands.HandleIncomingEbtReplicate) error
}

type HandlerEbtReplicate struct {
	handler EbtReplicateCommandHandler
}

func NewHandlerEbtReplicate(handler EbtReplicateCommandHandler) *HandlerEbtReplicate {
	return &HandlerEbtReplicate{handler: handler}
}

func (h HandlerEbtReplicate) Procedure() rpc.Procedure {
	return messages.EbtReplicateProcedure
}

func (h HandlerEbtReplicate) Handle(ctx context.Context, s mux.Stream, req *rpc.Request) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	args, err := messages.NewEbtReplicateArgumentsFromBytes(req.Arguments())
	if err != nil {
		return errors.Wrap(err, "error parsing arguments")
	}

	stream := ebt.NewIncomingStreamAdapter(s)

	cmd, err := commands.NewHandleIncomingEbtReplicate(args.Version(), args.Format(), stream)
	if err != nil {
		return errors.Wrap(err, "error creating the command")
	}

	err = h.handler.Handle(ctx, cmd)
	if err != nil {
		return errors.Wrap(err, "error executing the command")
	}

	return nil
}
