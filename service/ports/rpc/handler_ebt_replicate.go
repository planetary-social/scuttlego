package rpc

/*

type EbtReplicateCommandHandler interface {
	Handle(ctx context.Context, cmd commands.EbtReplicate) (<-chan messages.NoteOrMessage, error)
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
	cmd := commands.CreateWants{}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch, err := h.handler.Handle(ctx, cmd)
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

	return errors.New("not implemented")
}


*/
