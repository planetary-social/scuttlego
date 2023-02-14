package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type GetMessage struct {
	id refs.Message
}

func NewGetMessage(id refs.Message) (GetMessage, error) {
	if id.IsZero() {
		return GetMessage{}, errors.New("zero value of id")
	}
	return GetMessage{id: id}, nil
}

func (q *GetMessage) Id() refs.Message {
	return q.id
}

func (q *GetMessage) IsZero() bool {
	return q.id.IsZero()
}

type GetMessageHandler struct {
	transaction TransactionProvider
}

func NewGetMessageHandler(transaction TransactionProvider) *GetMessageHandler {
	return &GetMessageHandler{transaction: transaction}
}

func (h *GetMessageHandler) Handle(query GetMessage) (message.Message, error) {
	if query.IsZero() {
		return message.Message{}, errors.New("zero value of query")
	}

	var result message.Message
	if err := h.transaction.Transact(func(adapters Adapters) error {
		tmp, err := adapters.Message.Get(query.Id())
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
