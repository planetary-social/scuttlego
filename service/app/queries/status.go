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

type MessageRepository interface {
	// Count returns the number of stored messages.
	Count() (int, error)
}

type PeerManager interface {
	// Peers returns the currently connected peers.
	Peers() []transport.Peer
}

type StatusHandler struct {
	messageRepository MessageRepository
	feedRepository    FeedRepository
	peerManager       PeerManager
}

func NewStatusHandler(
	messageRepository MessageRepository,
	feedRepository FeedRepository,
	peerManager PeerManager,
) *StatusHandler {
	return &StatusHandler{
		messageRepository: messageRepository,
		feedRepository:    feedRepository,
		peerManager:       peerManager,
	}
}

func (h StatusHandler) Handle() (StatusResult, error) {
	numberOfMessages, err := h.messageRepository.Count()
	if err != nil {
		return StatusResult{}, errors.Wrap(err, "could not get the number of messages")
	}

	numberOfFeeds, err := h.feedRepository.Count()
	if err != nil {
		return StatusResult{}, errors.Wrap(err, "could not get the number of feeds")
	}

	peers, err := h.getPeers()
	if err != nil {
		return StatusResult{}, errors.Wrap(err, "could not get peers")
	}

	return StatusResult{
		NumberOfMessages: numberOfMessages,
		NumberOfFeeds:    numberOfFeeds,
		Peers:            peers,
	}, nil
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
