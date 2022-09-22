package ebt

import (
	"context"

	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
)

type CreateHistoryStreamHandler interface {
	Handle(ctx context.Context, query queries.CreateHistoryStream)
}

// todo rename
type StreamMessagesRequestHandler struct {
	handler CreateHistoryStreamHandler
}

func NewStreamMessagesRequestHandler(handler CreateHistoryStreamHandler) *StreamMessagesRequestHandler {
	return &StreamMessagesRequestHandler{handler: handler}
}

func (h StreamMessagesRequestHandler) Handle(ctx context.Context, id refs.Feed, seq *message.Sequence, messageWriter ebt.MessageWriter) {
	query := queries.CreateHistoryStream{
		Id:             id,
		Seq:            seq,
		Limit:          nil,
		Live:           true,
		Old:            true,
		ResponseWriter: NewMessageWriterAdapter(messageWriter),
	}

	h.handler.Handle(ctx, query)
}

type MessageWriterAdapter struct {
	messageWriter ebt.MessageWriter
}

func NewMessageWriterAdapter(messageWriter ebt.MessageWriter) *MessageWriterAdapter {
	return &MessageWriterAdapter{messageWriter: messageWriter}
}

func (m MessageWriterAdapter) WriteMessage(msg message.Message) error {
	return m.messageWriter.WriteMessage(msg)
}

func (m MessageWriterAdapter) CloseWithError(err error) error {
	return nil
}
