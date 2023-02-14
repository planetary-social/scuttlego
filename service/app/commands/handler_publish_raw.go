package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type PublishRaw struct {
	content []byte
}

func NewPublishRaw(content []byte) (PublishRaw, error) {
	if len(content) == 0 {
		return PublishRaw{}, errors.New("zero length of content")
	}

	return PublishRaw{content: content}, nil
}

func (cmd PublishRaw) IsZero() bool {
	return len(cmd.content) == 0
}

type PublishRawHandler struct {
	publisher RawMessagePublisher
	local     identity.Private
}

func NewPublishRawHandler(
	publisher RawMessagePublisher,
	local identity.Private,
) *PublishRawHandler {
	return &PublishRawHandler{
		publisher: publisher,
		local:     local,
	}
}

func (h *PublishRawHandler) Handle(cmd PublishRaw) (refs.Message, error) {
	content, err := message.NewRawContent(cmd.content)
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "could not create raw message content")
	}

	ref, err := h.publisher.Publish(h.local, content)
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "publishing error")
	}

	return ref, nil
}
