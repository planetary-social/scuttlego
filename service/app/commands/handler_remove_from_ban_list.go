package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/bans"
)

type RemoveFromBanList struct {
	hash bans.Hash
}

func NewRemoveFromBanList(hash bans.Hash) (RemoveFromBanList, error) {
	if hash.IsZero() {
		return RemoveFromBanList{}, errors.New("zero value of hash")
	}
	return RemoveFromBanList{hash: hash}, nil
}

func (c RemoveFromBanList) Hash() bans.Hash {
	return c.hash
}

func (c RemoveFromBanList) IsZero() bool {
	return c.hash.IsZero()
}

type RemoveFromBanListHandler struct {
	transaction TransactionProvider
}

func NewRemoveFromBanListHandler(transaction TransactionProvider) *RemoveFromBanListHandler {
	return &RemoveFromBanListHandler{
		transaction: transaction,
	}
}

func (h *RemoveFromBanListHandler) Handle(cmd RemoveFromBanList) error {
	if cmd.IsZero() {
		return errors.New("zero value of command")
	}

	if err := h.transaction.Transact(func(adapters Adapters) error {
		if err := adapters.BanList.Remove(cmd.Hash()); err != nil {
			return errors.Wrap(err, "could not remove the hash from the ban list")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
