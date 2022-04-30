package replication

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/refs"
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

const timeForReplicationLoop = 1 * time.Second

func (m Manager) sendFeedsToReplicateLoop(ctx context.Context, ch chan ReplicateFeedTask) {
	defer close(ch)

	// todo make this event driven
	// todo this doesn't take message buffer into account so there is some redundant work being done
	for {
		if err := m.sendFeedsToReplicate(ctx, ch); err != nil {
			m.logger.WithError(err).Error("could not send feeds to replicate")
		}

		select {
		case <-time.After(timeForReplicationLoop):
			continue
		case <-ctx.Done():
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
			m.logger.WithField("feed", task.Id).Debug("sent task")
			continue
		}
	}

	return nil
}
