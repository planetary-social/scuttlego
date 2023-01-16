package commands

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messagebuffer"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const (
	messageBufferPersistAtMessages = 1000
	messageBufferPersistEvery      = 5 * time.Second

	leaveUnpersistedMessagesFor = 15 * time.Second
)

type MessageBuffer struct {
	messages     messagesList
	messagesLock *sync.Mutex

	forcePersistCh   chan struct{}
	forcePersistOnce sync.Once

	transaction TransactionProvider
	logger      logging.Logger
}

func NewMessageBuffer(transaction TransactionProvider, logger logging.Logger) *MessageBuffer {
	return &MessageBuffer{
		messages:     make(messagesList),
		messagesLock: &sync.Mutex{},

		forcePersistCh: make(chan struct{}),

		transaction: transaction,
		logger:      logger.New("message_buffer"),
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

func (m *MessageBuffer) Handle(msg message.Message) error {
	m.messagesLock.Lock()
	defer m.messagesLock.Unlock()

	if err := m.messages.AddMessage(time.Now(), msg); err != nil {
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

	numberOfMessages := m.messages.MessageCount()
	if numberOfMessages == 0 {
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
		m.logger.WithField("key", key).WithField("sequence", updatedSequence.Int()).Trace("dropping after")
		m.messages[key].LeaveOnlyAfter(updatedSequence)
		if m.messages[key].Len() == 0 {
			m.logger.WithField("key", key).Trace("deleting")
			delete(m.messages, key)
		}
	}

	m.logger.
		WithField("count", numberOfMessages).
		WithField("duration", time.Since(start)).
		Debug("persisted messages")

	return nil
}

func (m *MessageBuffer) persistTransaction(adapters Adapters) (map[string]message.Sequence, error) {
	socialGraph, err := adapters.SocialGraph.GetSocialGraph()
	if err != nil {
		return nil, errors.Wrap(err, "could not load the social graph")
	}

	counterAll := 0
	counterSuccesses := 0

	updatedSequences := make(map[string]message.Sequence)

	for key, feedMessages := range m.messages {
		counterAll++

		feedRef := feedMessages.Feed()
		authorRef, err := refs.NewIdentityFromPublic(feedRef.Identity()) // todo cleanup
		if err != nil {
			return nil, errors.Wrap(err, "error creating an identity")
		}

		feedLogger := m.logger.WithField("feed", feedRef)

		feedIsBanned, err := adapters.BanList.ContainsFeed(feedRef)
		if err != nil {
			return nil, errors.Wrap(err, "error checking if the feed is banned")
		}

		if feedIsBanned {
			return nil, errors.New("feed is banned")
		}

		wantListContains, err := adapters.FeedWantList.Contains(feedRef)
		if err != nil {
			return nil, errors.Wrap(err, "error checking the want list")
		}

		if !socialGraph.HasContact(authorRef) && !wantListContains {
			continue // do nothing as this contact is not in our social graph
		}

		if err := adapters.Feed.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			seq := m.getSequence(feed)
			msgs := feedMessages.ConsecutiveSliceStartingWith(seq)

			feedLogger.
				WithField("sequence_in_database", seq.Int()).
				WithField("sequences_in_buffer", feedMessages.Sequences()).
				WithField("messages_that_can_be_persisted", len(msgs)).
				Trace("persisting messages")

			if len(msgs) > 0 {
				counterSuccesses++
			}

			for _, msg := range msgs {
				if err := feed.AppendMessage(msg); err != nil {
					// TODO if error then drop all messages?
					return errors.Wrap(err, "error updating feed")
				}
			}

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
		WithField("health", float64(counterSuccesses)/float64(counterAll)).
		WithField("messages", m.messages.MessageCount()).
		WithField("feeds", len(m.messages)).
		Debug("update complete")

	return updatedSequences, nil
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

func (f messagesList) AddMessage(t time.Time, msg message.Message) error {
	return f.get(msg.Feed()).Add(t, msg)
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
