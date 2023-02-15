package ebt

import (
	"context"
	"github.com/planetary-social/scuttlego/logging"

	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
)

type CreateHistoryStreamHandler interface {
	Handle(ctx context.Context, query queries.CreateHistoryStream)
}

type CreateHistoryStreamHandlerAdapter struct {
	handler CreateHistoryStreamHandler
	logger  logging.Logger
}

func NewCreateHistoryStreamHandlerAdapter(
	handler CreateHistoryStreamHandler,
	logger logging.Logger,
) *CreateHistoryStreamHandlerAdapter {
	return &CreateHistoryStreamHandlerAdapter{
		handler: handler,
		logger:  logger.New("create_history_stream_handler_adapter"),
	}
}

func (h CreateHistoryStreamHandlerAdapter) Handle(
	ctx context.Context,
	id refs.Feed,
	seq *message.Sequence,
	messageWriter ebt.MessageWriter,
) {
	query, err := queries.NewCreateHistoryStream(
		id,
		seq,
		nil,
		true,
		true,
		NewMessageWriterAdapter(messageWriter),
	)
	if err != nil {
		h.logger.WithError(err).Error("error creating query")
		return
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
