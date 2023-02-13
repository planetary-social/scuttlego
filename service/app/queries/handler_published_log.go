package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/common"
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
	feed        refs.Feed
	transaction TransactionProvider
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
		feed, err := adapters.Feed.GetFeed(h.feed)
		if err != nil {
			if errors.Is(err, common.ErrFeedNotFound) {
				return nil
			}
			return errors.Wrap(err, "could not get the feed")
		}

		messageSequence, ok := feed.Sequence()
		if !ok {
			return errors.New("we got a feed but the sequence was missing")
		}

		for {
			msg, err := adapters.Feed.GetMessage(h.feed, messageSequence)
			if err != nil {
				return errors.Wrap(err, "could not get the message")
			}

			receiveLogSequences, err := adapters.ReceiveLog.GetSequences(msg.Id())
			if err != nil {
				return errors.Wrap(err, "failed to look up message sequences")
			}

			receiveLogSequence := receiveLogSequences[0]

			if query.LastSeq != nil && query.LastSeq.Int() <= receiveLogSequence.Int() {
				break
			}

			result = append(result, LogMessage{
				Message:  msg,
				Sequence: receiveLogSequence,
			})

			tmp, previousSequenceExists := messageSequence.Previous()
			if !previousSequenceExists {
				break
			}

			messageSequence = tmp
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}
