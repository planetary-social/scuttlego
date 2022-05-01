package replication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/graph"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

const (
	delayIfNoFeedsToReplicate = 1 * time.Second

	backoffFriends = 30 * time.Second
	backoff        = 5 * time.Minute
	backoffFailed  = 10 * time.Minute

	refreshContactsEvery    = 5 * time.Second
	waitForTaskToBePickedUp = 1 * time.Second
)

type Storage interface {
	// GetContacts returns a list of contacts. Contacts must be sorted by hops,
	// ascending.
	GetContacts() ([]Contact, error)
}

type Contact struct {
	Who       refs.Feed
	Hops      graph.Hops
	FeedState FeedState
}

type Manager struct {
	storage Storage
	logger  logging.Logger

	activeTasks activeTasksMap
	peerState   peerMap
	lock        sync.Mutex

	contacts          []Contact
	contactsTimestamp time.Time
	contactsLock      sync.Mutex
}

func NewManager(logger logging.Logger, storage Storage) *Manager {
	return &Manager{
		storage:     storage,
		logger:      logger.New("manager"),
		activeTasks: make(activeTasksMap),
		peerState:   make(peerMap),
	}
}

func (m *Manager) GetFeedsToReplicate(ctx context.Context, remote identity.Public) <-chan ReplicateFeedTask {
	ch := make(chan ReplicateFeedTask)

	go m.sendFeedsToReplicateLoop(ctx, ch, remote)

	return ch
}

func (m *Manager) sendFeedsToReplicateLoop(ctx context.Context, ch chan ReplicateFeedTask, remote identity.Public) {
	defer close(ch)

	// todo make this event driven
	// todo this doesn't take message buffer into account so there is some redundant work being done
	for {
		if err := m.sendFeedToReplicate(ctx, ch, remote); err != nil {
			m.logger.WithError(err).Error("could not send feeds to replicate")
		}

		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}

func (m *Manager) sendFeedToReplicate(ctx context.Context, ch chan ReplicateFeedTask, remote identity.Public) error {
	contacts, err := m.getContacts()
	if err != nil {
		return errors.Wrap(err, "could not get contacts")
	}

	for i := range contacts {
		contact := contacts[i]

		shouldSendTask, err := m.startReplication(remote, contact)
		if err != nil {
			return errors.Wrap(err, "failed to start replication")
		}

		if shouldSendTask {
			task := ReplicateFeedTask{
				Id:    contact.Who,
				State: contact.FeedState,
				Ctx:   ctx,
				Complete: func(result TaskResult) {
					m.finishReplication(remote, contact, result)
				},
			}

			select {
			case <-ctx.Done():
				task.Complete(TaskResultDidNotStart)
				return ctx.Err()
			case <-time.After(waitForTaskToBePickedUp):
				task.Complete(TaskResultDidNotStart)
				return nil
			case ch <- task:
				return nil
			}
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delayIfNoFeedsToReplicate):
		return nil
	}
}

func (m *Manager) finishReplication(remote identity.Public, contact Contact, result TaskResult) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.logger.
		WithField("result", result).
		WithField("contact", contact.Who).
		Debug("finished replication")

	delete(m.activeTasks, contact.Who.String())

	if result != TaskResultDidNotStart {
		_, ok := m.peerState[remote.String()]
		if !ok {
			m.peerState[remote.String()] = make(peerState)
		}

		m.peerState[remote.String()][contact.Who.String()] = peerFeedState{
			LastReplicated: time.Now(),
			Result:         result,
		}
	}
}

func (m *Manager) startReplication(remote identity.Public, contact Contact) (bool, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	shouldStart, err := m.shouldStartReplication(remote, contact)
	if err != nil {
		return false, errors.Wrap(err, "error checking if replication should start")
	}

	if !shouldStart {
		return false, nil
	}

	m.activeTasks[contact.Who.String()] = struct{}{}
	return true, nil
}

func (m *Manager) shouldStartReplication(remote identity.Public, contact Contact) (bool, error) {
	_, ok := m.activeTasks[contact.Who.String()]
	if ok {
		return false, nil
	}

	peerState, ok := m.peerState[remote.String()]
	if !ok {
		return true, nil
	}

	peerFeedState, ok := peerState[contact.Who.String()]
	if !ok {
		return true, nil
	}

	switch peerFeedState.Result {
	case TaskResultHasMoreMessages:
		return true, nil
	case TaskResultFailed:
		return time.Since(peerFeedState.LastReplicated) > backoffFailed, nil
	case TaskResultDoesNotHaveMoreMessages:
		if contact.Hops.Int() > 1 {
			return time.Since(peerFeedState.LastReplicated) > backoff, nil
		} else {
			return time.Since(peerFeedState.LastReplicated) > backoffFriends, nil
		}
	default:
		return false, fmt.Errorf("unknown result '%v'", peerFeedState.Result)
	}
}

func (m *Manager) getContacts() ([]Contact, error) {
	m.contactsLock.Lock()
	defer m.contactsLock.Unlock()

	if time.Since(m.contactsTimestamp) > refreshContactsEvery {
		contacts, err := m.storage.GetContacts()
		if err != nil {
			return nil, errors.Wrap(err, "could not get contacts")
		}

		m.contacts = contacts
		m.contactsTimestamp = time.Now()
	}

	return m.contacts, nil
}

type peerMap map[string]peerState

type peerState map[string]peerFeedState

type peerFeedState struct {
	LastReplicated time.Time
	Result         TaskResult
}

type activeTasksMap map[string]struct{}
