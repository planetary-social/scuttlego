package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type PublishedLog struct {
	// Only messages with a sequence greater than the sequence of a message
	// pointed to by the given receive log sequence are returned. Given receive
	// log sequence must point to a message published by the current identity.
	// If not such message is known nil should be passed.
	LastSeq *common.ReceiveLogSequence
}

type PublishedLogHandler struct {
	transaction TransactionProvider
	feed        refs.Feed
}

func NewPublishedLogHandler(
	transaction TransactionProvider,
	local identity.Public,
) (*PublishedLogHandler, error) {
	localRef, err := refs.NewIdentityFromPublic(local)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a public identity")
	}

	return &PublishedLogHandler{
		transaction: transaction,
		feed:        localRef.MainFeed(),
	}, nil
}

func (h *PublishedLogHandler) Handle(query PublishedLog) ([]LogMessage, error) {
	var result []LogMessage

	if err := h.transaction.Transact(func(adapters Adapters) error {
		var startSeq *message.Sequence

		if query.LastSeq != nil {
			msg, err := adapters.ReceiveLog.GetMessage(*query.LastSeq)
			if err != nil {
				return errors.Wrap(err, "failed to find a message in the receive log")
			}

			if !msg.Feed().Equal(h.feed) {
				return errors.New("start sequence doesn't point to a message from this feed")
			}

			startSeq = internal.Ptr(msg.Sequence().Next())
		}

		msgs, err := adapters.Feed.GetMessages(h.feed, startSeq, nil)
		if err != nil {
			return errors.Wrap(err, "error getting messages")
		}

		for _, msg := range msgs {
			receiveLogSequences, err := adapters.ReceiveLog.GetSequences(msg.Id())
			if err != nil {
				return errors.Wrap(err, "failed to look up message sequences")
			}

			result = append(result, LogMessage{
				Message:  msg,
				Sequence: receiveLogSequences[0],
			})
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
