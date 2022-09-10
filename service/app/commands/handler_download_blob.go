package commands

import (
	"time"

	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const temporaryWantListDuration = 1 * time.Hour

type DownloadBlob struct {
	Id refs.Blob
}

type DownloadBlobHandler struct {
	transaction         TransactionProvider
	currentTimeProvider CurrentTimeProvider
}

func NewDownloadBlobHandler(
	transaction TransactionProvider,
	currentTimeProvider CurrentTimeProvider,
) *DownloadBlobHandler {
	return &DownloadBlobHandler{
		transaction:         transaction,
		currentTimeProvider: currentTimeProvider,
	}
}

func (h *DownloadBlobHandler) Handle(cmd DownloadBlob) error {
	until := h.currentTimeProvider.Get().Add(temporaryWantListDuration)
	return h.transaction.Transact(func(adapters Adapters) error {
		return adapters.WantList.Add(cmd.Id, until)
	})
}
