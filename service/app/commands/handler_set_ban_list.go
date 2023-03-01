package commands

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type SetBanList struct {
	hashes []bans.Hash
}

func NewSetBanList(hashes []bans.Hash) (SetBanList, error) {
	for _, hash := range hashes {
		if hash.IsZero() {
			return SetBanList{}, errors.New("zero value of hash")
		}
	}
	return SetBanList{hashes: hashes}, nil
}

type SetBanListHandler struct {
	transaction TransactionProvider
}

func NewSetBanListHandler(transaction TransactionProvider) *SetBanListHandler {
	return &SetBanListHandler{transaction: transaction}
}

func (h *SetBanListHandler) Handle(cmd SetBanList) error {
	if err := h.transaction.Transact(func(adapters Adapters) error {
		if err := adapters.BanList.Clear(); err != nil {
			return errors.Wrap(err, "error clearing the ban list")
		}

		for _, hash := range cmd.hashes {
			if err := adapters.BanList.Add(hash); err != nil {
				return errors.Wrap(err, "could not add the hash to the ban list")
			}

			if err := tryToRemoveTheBannedThing(adapters, hash); err != nil {
				return errors.Wrap(err, "error removing the banned thing")
			}
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func tryToRemoveTheBannedThing(adapters Adapters, hash bans.Hash) error {
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
