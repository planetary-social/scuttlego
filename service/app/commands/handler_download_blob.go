package commands

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const temporaryBlobWantListDuration = 1 * time.Hour

type DownloadBlob struct {
	id refs.Blob
}

func NewDownloadBlob(id refs.Blob) *DownloadBlob {
	return &DownloadBlob{id: id}
}

func (d DownloadBlob) Id() refs.Blob {
	return d.id
}

func (d DownloadBlob) IsZero() bool {
	return d.id.IsZero()
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
	if cmd.IsZero() {
		return errors.New("zero value of cmd")
	}

	until := h.currentTimeProvider.Get().Add(temporaryBlobWantListDuration)
	return h.transaction.Transact(func(adapters Adapters) error {
		return adapters.BlobWantList.Add(cmd.Id(), until)
	})
}
