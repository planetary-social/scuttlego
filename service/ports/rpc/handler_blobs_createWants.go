package rpc

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/app"
	"github.com/planetary-social/go-ssb/service/app/commands"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc/mux"
)

type HandlerBlobsCreateWants struct {
	app app.Application
}

func NewHandlerBlobsCreateWants(app app.Application) *HandlerBlobsCreateWants {
	return &HandlerBlobsCreateWants{
		app: app,
	}
}

func (h HandlerBlobsCreateWants) Procedure() rpc.Procedure {
	return messages.BlobsCreateWantsProcedure
}

func (h HandlerBlobsCreateWants) Handle(ctx context.Context, w mux.ResponseWriter, req *rpc.Request) error {
	cmd := commands.CreateWants{}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch, err := h.app.Commands.CreateWants.Handle(ctx, cmd)
	if err != nil {
		return errors.Wrap(err, "could not execute the create wants command")
	}

	for v := range ch {
		resp, err := messages.NewBlobsCreateWantsResponse(v.Id(), v.SizeOrWantDistance())
		if err != nil {
			return errors.Wrap(err, "failed to create a response")
		}

		j, err := resp.MarshalJSON()
		if err != nil {
			return errors.Wrap(err, "json marshalling failed")
		}

		if err := w.WriteMessage(j); err != nil {
			return errors.Wrap(err, "failed to send a message")
		}
	}

	return nil
}
