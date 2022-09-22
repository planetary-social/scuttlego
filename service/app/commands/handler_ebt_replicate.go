package commands

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
)

type HandleIncomingEbtReplicate struct {
	version int
	format  messages.EbtReplicateFormat
	stream  ebt.Stream
}

func NewHandleIncomingEbtReplicate(version int, format messages.EbtReplicateFormat, stream ebt.Stream) (HandleIncomingEbtReplicate, error) {
	if format.IsZero() {
		return HandleIncomingEbtReplicate{}, errors.New("zero value of format")
	}
	if stream == nil {
		return HandleIncomingEbtReplicate{}, errors.New("nil stream")
	}
	return HandleIncomingEbtReplicate{stream: stream}, nil
}

func (cmd HandleIncomingEbtReplicate) Version() int {
	return cmd.version
}

func (cmd HandleIncomingEbtReplicate) Format() messages.EbtReplicateFormat {
	return cmd.format
}

func (cmd HandleIncomingEbtReplicate) Stream() ebt.Stream {
	return cmd.stream
}

func (cmd HandleIncomingEbtReplicate) IsZero() bool {
	return cmd == HandleIncomingEbtReplicate{}
}

type HandleIncomingEbtReplicateHandler struct {
	replicator ebt.Replicator
}

func NewHandleIncomingEbtReplicateHandler(replicator ebt.Replicator) *HandleIncomingEbtReplicateHandler {
	return &HandleIncomingEbtReplicateHandler{replicator: replicator}
}

func (h *HandleIncomingEbtReplicateHandler) Handle(ctx context.Context, cmd HandleIncomingEbtReplicate) error {
	if cmd.IsZero() {
		return errors.New("zero value of command")
	}

	return h.replicator.HandleIncoming(ctx, cmd.Version(), cmd.Format(), cmd.Stream())
}
