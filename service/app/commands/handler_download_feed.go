package commands

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const temporaryFeedWantListDuration = 12 * time.Hour

type DownloadFeed struct {
	id refs.Feed
}

func NewDownloadFeed(id refs.Feed) (DownloadFeed, error) {
	if id.IsZero() {
		return DownloadFeed{}, errors.New("zero value of id")

	}
	return DownloadFeed{id: id}, nil
}

func (d DownloadFeed) Id() refs.Feed {
	return d.id
}

func (d DownloadFeed) IsZero() bool {
	return d.id.IsZero()
}

type DownloadFeedHandler struct {
	transaction         TransactionProvider
	currentTimeProvider CurrentTimeProvider
}

func NewDownloadFeedHandler(
	transaction TransactionProvider,
	currentTimeProvider CurrentTimeProvider,
) *DownloadFeedHandler {
	return &DownloadFeedHandler{
		transaction:         transaction,
		currentTimeProvider: currentTimeProvider,
	}
}

func (h *DownloadFeedHandler) Handle(cmd DownloadFeed) error {
	if cmd.IsZero() {
		return errors.New("zero value of command")
	}

	until := h.currentTimeProvider.Get().Add(temporaryFeedWantListDuration)
	return h.transaction.Transact(func(adapters Adapters) error {
		return adapters.FeedWantList.Add(cmd.Id(), until)
	})
}
