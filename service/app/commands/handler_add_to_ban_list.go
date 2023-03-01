package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/bans"
)

type AddToBanList struct {
	hash bans.Hash
}

func NewAddToBanList(hash bans.Hash) (AddToBanList, error) {
	if hash.IsZero() {
		return AddToBanList{}, errors.New("zero value of hash")
	}
	return AddToBanList{hash: hash}, nil
}

func (c AddToBanList) IsZero() bool {
	return c.hash.IsZero()
}

type AddToBanListHandler struct {
	transaction TransactionProvider
}

func NewAddToBanListHandler(
	transaction TransactionProvider,
) *AddToBanListHandler {
	return &AddToBanListHandler{
		transaction: transaction,
	}
}

func (h *AddToBanListHandler) Handle(cmd AddToBanList) error {
	if cmd.IsZero() {
		return errors.New("zero value of command")
	}

	if err := h.transaction.Transact(func(adapters Adapters) error {
		if err := adapters.BanList.Add(cmd.hash); err != nil {
			return errors.Wrap(err, "could not add the hash to the ban list")
		}
		return tryToRemoveTheBannedThing(adapters, cmd.hash)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
