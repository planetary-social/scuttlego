package queries

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type FeedRepository interface {
	// GetMessages returns messages with a sequence greater or equal to the provided sequence. If sequence is nil then
	// messages starting from the beginning of the feed are returned. Limit specifies the max number of returned
	// messages. If limit is nil then all messages matching the sequence criteria are returned.
	GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) // todo iterator instead of returning a huge array

	// Count returns the number of stored feeds.
	Count() (int, error)
}

type MessageSubscriber interface {
	// SubscribeToNewMessages subscribes to all new messages. Messages that belong to a specific feed are always
	// returned in order.
	SubscribeToNewMessages(ctx context.Context) <-chan message.Message
}

type CreateHistoryStream struct {
	Id refs.Feed

	// If not set messages starting from the very beginning of the feed will be returned. Otherwise, messages with
	// a sequence greater or equal to the provided one will be returned.
	// See: %ptQutWwkNIIteEn791Ru27DHtOsdnbcEJRgjuxW90Y4=.sha256
	Seq *message.Sequence

	// Number of messages to return, if not set then unlimited.
	Limit *int

	// If true then the channel will stay open and return further messages as they become available. This usually means
	// returning further messages which are replicated from other peers.
	Live bool

	// Used together with live. If true then old messages will be returned before serving live messages as they come in.
	// You most likely always want to set this to true. Setting Live to false and Old to false probably makes no sense
	// but won't cause an error, however no messages will be returned and the channel will be closed immediately.
	Old bool
}

type MessageWithErr struct {
	Message message.Message
	Err     error
}

type CreateHistoryStreamHandler struct {
	repository FeedRepository
	subscriber MessageSubscriber
}

func NewCreateHistoryStreamHandler(repository FeedRepository, subscriber MessageSubscriber) *CreateHistoryStreamHandler {
	return &CreateHistoryStreamHandler{
		repository: repository,
		subscriber: subscriber,
	}
}

// Handle returns a channel on which structs which contain messages and errors are received. The caller has to check
// the error. The message can be used only if the error is nil. If an error is not nil then no further values
// will be sent; the channel will be closed. It is possible that no values will be returned and the channel will be
// closed immediately. If the context terminates an error will be returned.
func (h *CreateHistoryStreamHandler) Handle(ctx context.Context, query CreateHistoryStream) <-chan MessageWithErr {
	ch := make(chan MessageWithErr)
	go h.sendMessagesOrError(ctx, query, ch)
	return ch
}

func (h *CreateHistoryStreamHandler) sendMessagesOrError(ctx context.Context, query CreateHistoryStream, ch chan<- MessageWithErr) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer close(ch)

	if err := h.sendMessages(ctx, query, ch); err != nil {
		ch <- MessageWithErr{Err: errors.Wrap(err, "could not send messages")}
	}
}

func (h *CreateHistoryStreamHandler) sendMessages(ctx context.Context, query CreateHistoryStream, ch chan<- MessageWithErr) error {
	lastSequence := query.Seq
	sentMessages := 0

	var liveMsgs <-chan message.Message
	if query.Live {
		liveMsgs = h.subscriber.SubscribeToNewMessages(ctx)
	}

	if query.Old {
		msgs, err := h.repository.GetMessages(query.Id, query.Seq, query.Limit)
		if err != nil {
			return errors.Wrap(err, "could not retrieve messages")
		}

		for _, msg := range msgs {
			select {
			case ch <- MessageWithErr{Message: msg}:
				seq := msg.Sequence()
				lastSequence = &seq
				sentMessages++
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	if query.Live {
		for {
			if query.Limit != nil && sentMessages >= *query.Limit {
				return nil
			}

			msg, ok := <-liveMsgs
			if !ok {
				return errors.New("live messages channel closed")
			}

			if !msg.Feed().Equal(query.Id) {
				continue
			}

			if lastSequence != nil && !msg.Sequence().ComesAfter(*lastSequence) {
				continue
			}

			select {
			case ch <- MessageWithErr{Message: msg}:
				seq := msg.Sequence()
				lastSequence = &seq
				sentMessages++
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}
