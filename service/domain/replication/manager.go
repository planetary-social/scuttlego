package replication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/graph"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

const (
	// If at some point there are no feeds to replicate then the manager will
	// wait for this duration before reattempting to find a feed which could be
	// replicated by the particular peer. This is done to avoid unnecessary
	// busy-looping.
	delayIfNoFeedsToReplicate = 1 * time.Second

	// Specifies for how long the manager will avoid asking the peer for
	// messages from a feed which is 1 or fewer hops away after the peer reports
	// that it has no new messages for that feed.
	backoffFriends = 30 * time.Second

	// Specifies for how long the manager will avoid asking the peer for
	// messages from a feed which is more than 1 hop away after the peer reports
	// that it has no new messages for that feed.
	backoff = 5 * time.Minute

	// Specifies for how long the manager will avoid asking the peer for
	// messages from a feed if a replication task fails eg. an invalid message
	// is returned by the peer.
	backoffFailed = 10 * time.Minute

	// For how long the social graph will be cached before rebuilding it. Until
	// this refresh happens newly discovered feeds are not taken into account.
	refreshContactsEvery = 5 * time.Second

	// For how long to wait for the task to be picked up by a peer's replicator
	// before giving up and making that feed available to be replicated by other
	// peers.
	waitForTaskToBePickedUp = 10 * time.Millisecond
)

type Storage interface {
	// GetContacts returns a list of contacts. Contacts are sorted by hops,
	// ascending. Contacts include the local feed.
	GetContacts() ([]Contact, error)
}

type MessageBuffer interface {
	Sequence(feed refs.Feed) (message.Sequence, bool)
}

type Contact struct {
	Who       refs.Feed
	Hops      graph.Hops
	FeedState FeedState
}

// Manager distributes replication tasks to replicators. Replicators consume
// those tasks to run replication processes on connected peers.
//
// For better fine-grained control of which feeds are being replicated the
// replicators shouldn't utilize live replication.
//
// During replication feeds which have a lower number of hops to the local
// identity are prioritized. Only one peer will be asked to replicate a
// particular feed at any given time. Manager backs off if a peer doesn't have
// any new messages for a feed before attempting to ask it for messages from
// that feed again. Backoff time is increased for feeds which are further away.
type Manager struct {
	storage Storage
	buffer  MessageBuffer
	logger  logging.Logger

	activeTasks *activeTasksSet
	peerState   peerMap    // todo clean up periodically
	lock        sync.Mutex // locks activeTasks and peerState

	contactsCache          []Contact
	contactsCacheTimestamp time.Time
	contactsCacheLock      sync.Mutex // locks contactsCache and contactsCacheTimestamp
}

func NewManager(logger logging.Logger, storage Storage, buffer MessageBuffer) *Manager {
	return &Manager{
		storage:     storage,
		buffer:      buffer,
		logger:      logger.New("manager"),
		activeTasks: newActiveTasksSet(),
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

	for {
		if err := m.sendFeedToReplicate(ctx, ch, remote); err != nil {
			m.logger.WithError(err).Error("send feed to replicate failed")
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

		state, err := m.bufferState(contact)
		if err != nil {
			return errors.Wrap(err, "error determining feed state")
		}

		if shouldSendTask {
			task := ReplicateFeedTask{
				Id:    contact.Who,
				State: state,
				Ctx:   ctx,
				OnComplete: func(result TaskResult) {
					m.finishReplication(remote, contact, result)
				},
			}

			select {
			case <-ctx.Done():
				task.OnComplete(TaskResultDidNotStart)
				return ctx.Err()
			case <-time.After(waitForTaskToBePickedUp):
				task.OnComplete(TaskResultDidNotStart)
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

func (m *Manager) bufferState(contact Contact) (FeedState, error) {
	bufferSequence, inBuffer := m.buffer.Sequence(contact.Who)
	storageSequence, inStorage := contact.FeedState.Sequence()

	if !inBuffer && !inStorage {
		return NewEmptyFeedState(), nil
	}

	if inBuffer && !inStorage {
		return NewFeedState(bufferSequence)
	}

	if !inBuffer && inStorage {
		return NewFeedState(storageSequence)
	}

	if bufferSequence.ComesAfter(storageSequence) {
		return NewFeedState(bufferSequence)
	} else {
		return NewFeedState(storageSequence)
	}
}

func (m *Manager) finishReplication(remote identity.Public, contact Contact, result TaskResult) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.logger.
		WithField("result", result).
		WithField("contact", contact.Who).
		Trace("finished replication")

	m.activeTasks.Delete(contact.Who)

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

	m.activeTasks.Put(contact.Who)
	return true, nil
}

func (m *Manager) shouldStartReplication(remote identity.Public, contact Contact) (bool, error) {
	if m.activeTasks.Contains(contact.Who) {
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
	m.contactsCacheLock.Lock()
	defer m.contactsCacheLock.Unlock()

	if time.Since(m.contactsCacheTimestamp) > refreshContactsEvery {
		contacts, err := m.storage.GetContacts()
		if err != nil {
			return nil, errors.Wrap(err, "could not get fresh contacts")
		}

		m.contactsCache = contacts
		m.contactsCacheTimestamp = time.Now()
	}

	return m.contactsCache, nil
}

type peerMap map[string]peerState

type peerState map[string]peerFeedState

type peerFeedState struct {
	LastReplicated time.Time
	Result         TaskResult
}

type activeTasksSet struct {
	m map[string]struct{}
}

func newActiveTasksSet() *activeTasksSet {
	return &activeTasksSet{
		m: make(map[string]struct{}),
	}
}

func (s *activeTasksSet) Contains(feed refs.Feed) bool {
	_, ok := s.m[feed.String()]
	return ok
}

func (s *activeTasksSet) Put(feed refs.Feed) {
	s.m[feed.String()] = struct{}{}
}

func (s *activeTasksSet) Delete(feed refs.Feed) {
	delete(s.m, feed.String())
}
