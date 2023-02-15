package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type RawMessagePublisher interface {
	Publish(identity identity.Private, content message.RawContent) (refs.Message, error)
}

type PublishRawAsIdentity struct {
	content  []byte
	identity identity.Private
}

func NewPublishRawAsIdentity(content []byte, identity identity.Private) (PublishRawAsIdentity, error) {
	if len(content) == 0 {
		return PublishRawAsIdentity{}, errors.New("zero length of content")
	}

	if identity.IsZero() {
		return PublishRawAsIdentity{}, errors.New("zero value of identity")
	}

	return PublishRawAsIdentity{content: content, identity: identity}, nil
}

func (cmd PublishRawAsIdentity) Content() []byte {
	return cmd.content
}

func (cmd PublishRawAsIdentity) Identity() identity.Private {
	return cmd.identity
}

func (cmd PublishRawAsIdentity) IsZero() bool {
	return cmd.identity.IsZero()
}

type PublishRawAsIdentityHandler struct {
	publisher RawMessagePublisher
}

func NewPublishRawAsIdentityHandler(
	publisher RawMessagePublisher,
) *PublishRawAsIdentityHandler {
	return &PublishRawAsIdentityHandler{
		publisher: publisher,
	}
}

func (h *PublishRawAsIdentityHandler) Handle(cmd PublishRawAsIdentity) (refs.Message, error) {
	if cmd.IsZero() {
		return refs.Message{}, errors.New("zero value of cmd")
	}

	content, err := message.NewRawContent(cmd.Content())
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "could not create raw message content")
	}

	ref, err := h.publisher.Publish(cmd.Identity(), content)
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "publishing error")
	}

	return ref, nil
}
