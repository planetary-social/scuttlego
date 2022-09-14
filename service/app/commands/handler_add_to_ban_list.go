package commands

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/refs"
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
		return h.tryToRemoveTheBannedThing(adapters, cmd.hash)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func (h *AddToBanListHandler) tryToRemoveTheBannedThing(adapters Adapters, hash bans.Hash) error {
	bannableRef, err := adapters.BanList.LookupMapping(hash)
	if err != nil {
		if errors.Is(err, ErrBanListMappingNotFound) {
			return nil
		}
		return errors.Wrap(err, "error looking up bannable refs")
	}

	switch v := bannableRef.Value().(type) {
	case refs.Feed:
		if err := adapters.Feed.DeleteFeed(v); err != nil {
			return errors.Wrap(err, "failed to delete a feed")
		}
		return nil
	default:
		return fmt.Errorf("unknown bannable ref type '%T'", bannableRef.Value())
	}
}
