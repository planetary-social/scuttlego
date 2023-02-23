package commands

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/messagebuffer"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const (
	messageBufferPersistAtMessages = 250
	messageBufferPersistEvery      = 5 * time.Second

	leaveUnpersistedMessagesFor = 15 * time.Second
)

type RawMessageIdentifier interface {
	PeekRawMessage(raw message.RawMessage) (feeds.PeekedMessage, error)
	VerifyRawMessage(raw message.RawMessage) (message.Message, error)
}

type ForkedFeedTracker interface {
	AddForkedFeed(replicatedFrom identity.Public, feed refs.Feed)
}

type MessageBuffer struct {
	messages     messagesList
	messagesLock *sync.Mutex

	forcePersistCh   chan struct{}
	forcePersistOnce sync.Once

	transaction       TransactionProvider
	identifier        RawMessageIdentifier
	forkedFeedTracker ForkedFeedTracker
	logger            logging.Logger
}

func NewMessageBuffer(
	transaction TransactionProvider,
	identifier RawMessageIdentifier,
	forkedFeedTracker ForkedFeedTracker,
	logger logging.Logger,
) *MessageBuffer {
	return &MessageBuffer{
		messages:     make(messagesList),
		messagesLock: &sync.Mutex{},

		forcePersistCh: make(chan struct{}),

		transaction:       transaction,
		identifier:        identifier,
		forkedFeedTracker: forkedFeedTracker,
		logger:            logger.New("message_buffer"),
	}
}

func (m *MessageBuffer) Run(ctx context.Context) error {
	for {
		if err := m.persist(); err != nil {
			m.logger.WithError(err).Error("error persisting messages")
		}

		m.cleanup()

		select {
		case <-time.After(messageBufferPersistEvery):
			continue
		case <-m.forcePersistCh:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *MessageBuffer) Handle(replicatedFrom identity.Public, rawMsg message.RawMessage) error {
	m.messagesLock.Lock()
	defer m.messagesLock.Unlock()

	msg, err := m.identifier.PeekRawMessage(rawMsg)
	if err != nil {
		return errors.Wrap(err, "failed to identify the raw message")
	}

	rm, err := messagebuffer.NewReceivedMessage[feeds.PeekedMessage](replicatedFrom, msg)
	if err != nil {
		return errors.Wrap(err, "error creating received message")
	}

	if err := m.messages.AddMessage(time.Now(), rm); err != nil {
		return errors.Wrap(err, "could not add a message")
	}

	if m.messages.MessageCount() > messageBufferPersistAtMessages {
		m.forcePersistOnce.Do(
			func() {
				close(m.forcePersistCh)
			},
		)
	}

	return nil
}

func (m *MessageBuffer) persist() error {
	m.messagesLock.Lock()
	defer m.messagesLock.Unlock()

	defer func() {
		m.forcePersistCh = make(chan struct{})
		m.forcePersistOnce = sync.Once{}
	}()

	if m.messages.MessageCount() == 0 {
		return nil
	}

	start := time.Now()

	var updatedSequences map[string]message.Sequence

	if err := m.transaction.Transact(func(adapters Adapters) (err error) {
		updatedSequences, err = m.persistTransaction(adapters)
		return err
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	for key, updatedSequence := range updatedSequences {
		logger := m.logger.WithField("key", key).WithField("updated_sequence", updatedSequence.Int())

		logger.Trace("leaving only after")
		m.messages[key].LeaveOnlyAfter(updatedSequence)

		if m.messages[key].Len() == 0 {
			logger.Trace("deleting")
			delete(m.messages, key)
		}
	}

	m.logger.
		WithField("duration", time.Since(start)).
		Debug("persisted messages")

	return nil
}

func (m *MessageBuffer) persistTransaction(adapters Adapters) (map[string]message.Sequence, error) {
	socialGraph, err := adapters.SocialGraph.GetSocialGraph()
	if err != nil {
		return nil, errors.Wrap(err, "could not load the social graph")
	}

	counterAllMessages := 0
	counterConsideredMessages := 0
	counterPersistedMessages := 0

	updatedSequences := make(map[string]message.Sequence)

	for key, feedMessages := range m.messages {
		counterAllMessages += feedMessages.Len()

		feedRef := feedMessages.Feed()
		feedLogger := m.logger.WithField("feed", feedRef)

		shouldSave, err := m.shouldSave(adapters, socialGraph, feedRef)
		if err != nil {
			return nil, errors.Wrap(err, "error checking if this feed should be saved")
		}

		if !shouldSave {
			delete(m.messages, key)
			continue
		}

		if err := adapters.Feed.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			seq := m.getSequence(feed)
			msgs := feedMessages.ConsecutiveSliceStartingWith(seq)
			counterConsideredMessages += len(msgs)

			var validatedMessages []messagebuffer.ReceivedMessage[message.Message]
			for _, peekedMessage := range msgs {
				msg, err := m.identifier.VerifyRawMessage(peekedMessage.Message().Raw())
				if err != nil {
					feedLogger.WithError(err).Error("error verifying message")
					continue
				}

				receivedMsg, err := messagebuffer.NewReceivedMessage(peekedMessage.ReplicatedFrom(), msg)
				if err != nil {
					return errors.Wrap(err, "error creating received message")
				}

				validatedMessages = append(validatedMessages, receivedMsg)
			}

			feedLogger.
				WithField("sequence_in_database", seq).
				WithField("sequences_in_buffer", feedMessages.Sequences()).
				WithField("messages_considered_for_persisting", len(msgs)).
				Trace("persisting messages")

			for _, msg := range validatedMessages {
				if err := feed.AppendMessage(msg.Message()); err != nil {
					// probably a forked feed
					feedLogger.
						WithError(err).
						WithField("message_being_appended", msg).
						Trace("error appending message")
					feedMessages.Remove(msg.Message().Raw()) // todo?
					m.forkedFeedTracker.AddForkedFeed(msg.ReplicatedFrom(), msg.Message().Feed())
					return nil
				}
			}

			counterPersistedMessages += len(feed.MessagesThatWillBePersisted())

			sequence, ok := feed.Sequence()
			if ok {
				updatedSequences[key] = sequence
			}

			return nil
		}); err != nil {
			return nil, errors.Wrapf(err, "failed to update the feed '%s'", feedRef)
		}
	}

	m.logger.
		WithField("remaining_messages", m.messages.MessageCount()).
		WithField("remaining_feeds", len(m.messages)).
		WithField("messages_all", counterAllMessages).
		WithField("messages_considered", counterConsideredMessages).
		WithField("messages_persisted", counterPersistedMessages).
		WithField("health", float64(counterPersistedMessages)/float64(counterAllMessages)).
		Debug("update complete")

	return updatedSequences, nil
}

func (m *MessageBuffer) shouldSave(adapters Adapters, socialGraph graph.SocialGraph, feedRef refs.Feed) (bool, error) {
	authorRef, err := refs.NewIdentityFromPublic(feedRef.Identity()) // todo cleanup
	if err != nil {
		return false, errors.Wrap(err, "error creating an identity")
	}

	feedIsBanned, err := adapters.BanList.ContainsFeed(feedRef)
	if err != nil {
		return false, errors.Wrap(err, "error checking if the feed is banned")
	}

	if feedIsBanned {
		return false, nil
	}

	wantListContains, err := adapters.FeedWantList.Contains(feedRef)
	if err != nil {
		return false, errors.Wrap(err, "error checking the want list")
	}

	if !socialGraph.HasContact(authorRef) && !wantListContains {
		return false, nil
	}

	return true, nil
}

func (m *MessageBuffer) cleanup() {
	m.messagesLock.Lock()
	defer m.messagesLock.Unlock()

	messagesBefore := m.messages.MessageCount()
	feedsBefore := len(m.messages)

	for key, feedMessages := range m.messages {
		feedMessages.RemoveOlderThan(time.Now().Add(-leaveUnpersistedMessagesFor))
		if feedMessages.Len() == 0 {
			delete(m.messages, key)
		}
	}

	messagesAfter := m.messages.MessageCount()
	feedsAfter := len(m.messages)

	m.logger.
		WithField("messages_before", messagesBefore).
		WithField("messages_after", messagesAfter).
		WithField("feeds_before", feedsBefore).
		WithField("feeds_after", feedsAfter).
		Trace("cleanup complete")
}

func (m *MessageBuffer) getSequence(feed *feeds.Feed) *message.Sequence {
	seq, ok := feed.Sequence()
	if !ok {
		return nil
	}
	return &seq
}

type messagesList map[string]*messagebuffer.FeedMessages

func (f messagesList) AddMessage(t time.Time, rm messagebuffer.ReceivedMessage[feeds.PeekedMessage]) error {
	return f.get(rm.Message().Feed()).Add(t, rm)
}

func (f messagesList) get(feed refs.Feed) *messagebuffer.FeedMessages {
	key := feed.String()
	v, ok := f[key]
	if !ok {
		v = messagebuffer.NewFeedMessages(feed)
		f[key] = v
	}
	return v
}

func (f messagesList) MessageCount() int {
	var result int
	for _, messages := range f {
		result += messages.Len()
	}
	return result
}
