package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type StatusResult struct {
	NumberOfMessages int
	NumberOfFeeds    int
	Peers            []Peer
}

type Peer struct {
	Identity identity.Public
}

type PeerManager interface {
	// Peers returns the currently connected peers.
	Peers() []transport.Peer
}

type StatusHandler struct {
	transaction TransactionProvider
	peerManager PeerManager
}

func NewStatusHandler(
	transaction TransactionProvider,
	peerManager PeerManager,
) *StatusHandler {
	return &StatusHandler{
		transaction: transaction,
		peerManager: peerManager,
	}
}

func (h StatusHandler) Handle() (StatusResult, error) {
	var result StatusResult

	if err := h.transaction.Transact(func(adapters Adapters) error {
		numberOfFeeds, err := adapters.Feed.Count()
		if err != nil {
			return errors.Wrap(err, "could not get the number of feeds")
		}

		numberOfMessages, err := adapters.Message.Count()
		if err != nil {
			return errors.Wrap(err, "could not get the number of messages")
		}

		result.NumberOfFeeds = numberOfFeeds
		result.NumberOfMessages = numberOfMessages
		return nil
	}); err != nil {
		return StatusResult{}, errors.Wrap(err, "transaction failed")
	}

	peers, err := h.getPeers()
	if err != nil {
		return StatusResult{}, errors.Wrap(err, "could not get peers")
	}

	result.Peers = peers
	return result, nil
}

func (h StatusHandler) getPeers() ([]Peer, error) {
	var result []Peer

	for _, peer := range h.peerManager.Peers() {
		result = append(result, Peer{
			Identity: peer.Identity(),
		})
	}

	return result, nil
}
