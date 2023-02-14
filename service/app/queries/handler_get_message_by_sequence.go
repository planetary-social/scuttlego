package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type GetMessageBySequence struct {
	feed     refs.Feed
	sequence message.Sequence
}

func NewGetMessageBySequence(feed refs.Feed, sequence message.Sequence) (GetMessageBySequence, error) {
	if feed.IsZero() {
		return GetMessageBySequence{}, errors.New("zero value of feed")
	}
	if sequence.IsZero() {
		return GetMessageBySequence{}, errors.New("zero value of sequence")
	}
	return GetMessageBySequence{feed: feed, sequence: sequence}, nil
}

func (q *GetMessageBySequence) Feed() refs.Feed {
	return q.feed
}

func (q *GetMessageBySequence) Sequence() message.Sequence {
	return q.sequence
}

func (q *GetMessageBySequence) IsZero() bool {
	return q.feed.IsZero()
}

type GetMessageBySequenceHandler struct {
	transaction TransactionProvider
}

func NewGetMessageBySequenceHandler(transaction TransactionProvider) *GetMessageBySequenceHandler {
	return &GetMessageBySequenceHandler{transaction: transaction}
}

func (h *GetMessageBySequenceHandler) Handle(query GetMessageBySequence) (message.Message, error) {
	if query.IsZero() {
		return message.Message{}, errors.New("zero value of query")
	}

	var result message.Message
	if err := h.transaction.Transact(func(adapters Adapters) error {
		tmp, err := adapters.Feed.GetMessage(query.Feed(), query.Sequence())
		if err != nil {
			return errors.Wrap(err, "error getting message")
		}
		result = tmp
		return nil
	}); err != nil {
		return message.Message{}, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
