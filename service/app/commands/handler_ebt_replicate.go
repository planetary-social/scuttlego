package commands

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/messages"
)

type EbtReplicate struct {
	ch <-chan messages.NoteOrMessage
}

func NewEbtReplicate(ch <-chan messages.NoteOrMessage) (EbtReplicate, error) {
	if ch == nil {
		return EbtReplicate{}, errors.New("channel is not initialized")
	}
	return EbtReplicate{ch: ch}, nil
}

func (cmd EbtReplicate) Ch() <-chan messages.NoteOrMessage {
	return cmd.ch
}

func (cmd EbtReplicate) IsZero() bool {
	return cmd == EbtReplicate{}
}

type EbtReplicateHandler struct {
}

func NewEbtReplicateHandler() *EbtReplicateHandler {
	return &EbtReplicateHandler{}
}

func (h *EbtReplicateHandler) Handle(ctx context.Context, cmd EbtReplicate) (<-chan messages.NoteOrMessage, error) {
	if cmd.IsZero() {
		return nil, errors.New("zero value of command")
	}

	return nil, errors.New("not implemented")
}
