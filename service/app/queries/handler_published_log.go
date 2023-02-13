package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type PublishedLog struct {
	// Only messages published by the current identity with receive log sequence
	// numbers greater than the given receive log sequence are returned. If a
	// message has multiple receive log sequence numbers the higher one is used.
	// Pass nil to get all messages.
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

			receiveLogSequence, err := h.highestReceiveLogSequence(receiveLogSequences)
			if err != nil {
				return errors.Wrap(err, "error getting highest receive log sequence")
			}

			if query.LastSeq != nil && query.LastSeq.Int() >= receiveLogSequence.Int() {
				break
			}

			result = append(
				[]LogMessage{
					{
						Message:  msg,
						Sequence: receiveLogSequence,
					},
				}, result...)

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

func (h *PublishedLogHandler) highestReceiveLogSequence(sequences []common.ReceiveLogSequence) (common.ReceiveLogSequence, error) {
	if len(sequences) == 0 {
		return common.ReceiveLogSequence{}, errors.New("no sequences given")
	}

	var result *common.ReceiveLogSequence
	for i := range sequences {
		if result == nil || sequences[i].Int() > result.Int() {
			result = &sequences[i]
		}
	}
	return *result, nil
}
