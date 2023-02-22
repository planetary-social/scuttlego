// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package tests

import (
	"testing"

	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/replication/gossip"
)

// Injectors from wire.go:

func BuildTestReplication(t *testing.T) (TestReplication, error) {
	devNullLogger := logging.NewDevNullLogger()
	sessionTracker := ebt.NewSessionTracker()
	rawMessageHandlerMock := NewRawMessageHandlerMock()
	wantedFeedsProviderMock := NewWantedFeedsProviderMock()
	wantedFeedsCache := replication.NewWantedFeedsCache(wantedFeedsProviderMock)
	messageStreamerMock := NewMessageStreamerMock()
	sessionRunner := ebt.NewSessionRunner(devNullLogger, rawMessageHandlerMock, wantedFeedsCache, messageStreamerMock)
	manager := gossip.NewManager(devNullLogger, wantedFeedsCache)
	gossipReplicator, err := gossip.NewGossipReplicator(manager, rawMessageHandlerMock, devNullLogger)
	if err != nil {
		return TestReplication{}, err
	}
	replicator := ebt.NewReplicator(sessionTracker, sessionRunner, gossipReplicator, devNullLogger)
	negotiator := replication.NewNegotiator(devNullLogger, replicator, gossipReplicator)
	testReplication := TestReplication{
		Negotiator:         negotiator,
		RawMessageHandler:  rawMessageHandlerMock,
		ContactsRepository: wantedFeedsProviderMock,
		MessageStreamer:    messageStreamerMock,
	}
	return testReplication, nil
}

// wire.go:

type TestReplication struct {
	Negotiator *replication.Negotiator

	RawMessageHandler  *RawMessageHandlerMock
	ContactsRepository *WantedFeedsProviderMock
	MessageStreamer    *MessageStreamerMock
}
