package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
)

type RawMessageHandler struct {
	identifier RawMessageIdentifier
	buffer     *MessageBuffer
	logger     logging.Logger
}

func NewRawMessageHandler(
	identifier RawMessageIdentifier,
	buffer *MessageBuffer,
	logger logging.Logger,
) *RawMessageHandler {
	return &RawMessageHandler{
		identifier: identifier,
		buffer:     buffer,
		logger:     logger.New("raw_message_handler"),
	}
}

func (h *RawMessageHandler) Handle(replicatedFrom identity.Public, rawMsg message.RawMessage) error {
	if err := h.buffer.Handle(replicatedFrom, rawMsg); err != nil {
		return errors.Wrap(err, "failed to put the message in the buffer")
	}
	return nil
}
