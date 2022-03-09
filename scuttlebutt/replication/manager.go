package replication

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/refs"
)

var ErrFeedNotFound = errors.New("feed not found")

type Storage interface {
	GetContacts() ([]Contact, error)
}

type Contact struct {
	Who       refs.Feed
	FeedState FeedState
}

type Manager struct {
	storage Storage
	logger  logging.Logger
}

func NewManager(logger logging.Logger, storage Storage) *Manager {
	return &Manager{
		storage: storage,
		logger:  logger.New("manager"),
	}
}

func (m Manager) GetFeedsToReplicate(ctx context.Context) <-chan ReplicateFeedTask {
	ch := make(chan ReplicateFeedTask)

	go m.sendFeedsToReplicateLoop(ctx, ch)

	return ch
}

const timeForReplicationLoop = 60 * time.Second

func (m Manager) sendFeedsToReplicateLoop(ctx context.Context, ch chan ReplicateFeedTask) {
	// todo make this event driven
	for {
		batchCtx, cancel := context.WithTimeout(ctx, timeForReplicationLoop)

		if err := m.sendFeedsToReplicate(batchCtx, ch); err != nil {
			m.logger.WithError(err).Error("could not send feeds to replicate")
		}

		select {
		case <-batchCtx.Done():
			cancel()
			continue
		case <-ctx.Done():
			cancel()
			return
		}
	}
}

func (m Manager) sendFeedsToReplicate(ctx context.Context, ch chan ReplicateFeedTask) error {
	contacts, err := m.storage.GetContacts()
	if err != nil {
		return errors.Wrap(err, "could not get contacts")
	}

	m.logger.WithField("n", len(contacts)).Debug("got contacts to replicate")

	for _, contact := range contacts {
		task := ReplicateFeedTask{
			Id:    contact.Who,
			State: contact.FeedState,
			Ctx:   ctx,
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- task:
			continue
		}
	}

	return nil
}
