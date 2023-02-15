package commands

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/messages"
)

type BlobReplicationManager interface {
	HandleIncomingCreateWantsRequest(ctx context.Context) (<-chan messages.BlobWithSizeOrWantDistance, error)
}

type CreateWantsHandler struct {
	manager BlobReplicationManager
}

func NewCreateWantsHandler(manager BlobReplicationManager) *CreateWantsHandler {
	return &CreateWantsHandler{
		manager: manager,
	}
}

func (h *CreateWantsHandler) Handle(ctx context.Context) (<-chan messages.BlobWithSizeOrWantDistance, error) {
	return h.manager.HandleIncomingCreateWantsRequest(ctx)
}
