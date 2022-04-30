package commands

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/feeds"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
)

const (
	messageBufferMaxMessages  = 1000
	messageBufferPersistEvery = 5 * time.Second
)

type MessageBuffer struct {
	feeds          feedsMap
	feedsLock      *sync.Mutex
	forcePersistCh chan struct{}
	transaction    TransactionProvider
	logger         logging.Logger
}

func NewMessageBuffer(transaction TransactionProvider, logger logging.Logger) *MessageBuffer {
	return &MessageBuffer{
		feeds:          make(feedsMap),
		feedsLock:      &sync.Mutex{},
		forcePersistCh: make(chan struct{}),
		transaction:    transaction,
		logger:         logger.New("message_buffer"),
	}
}

func (m *MessageBuffer) Run(ctx context.Context) error {
	for {
		if err := m.persist(); err != nil {
			m.logger.WithError(err).Error("error persisting messages")
		}

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
	m.feedsLock.Lock()
	defer m.feedsLock.Unlock()

	m.logger.
		WithField("sequence", msg.Sequence().Int()).
		WithField("feed", msg.Feed().String()).
		Trace("handling a new message")

	m.feeds.AddMessage(msg)
	m.maybeTriggerPersist()
	return nil
}

func (m *MessageBuffer) maybeTriggerPersist() {
	if m.feeds.MessageCount() < messageBufferMaxMessages {
		return
	}

	go func() {
		select {
		case m.forcePersistCh <- struct{}{}:
		case <-time.After(10 * time.Millisecond):
		}
	}()
}

func (m *MessageBuffer) persist() error {
	m.feedsLock.Lock()
	defer m.feedsLock.Unlock()

	numberOfMessages := m.feeds.MessageCount()

	if numberOfMessages == 0 {
		return nil
	}

	start := time.Now()

	if err := m.transaction.Transact(func(adapters Adapters) error {
		return m.persistAll(adapters)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	m.logger.
		WithField("messages", numberOfMessages).
		WithField("duration", time.Since(start)).
		Debug("persisted messages")

	return nil
}

func (m *MessageBuffer) persistAll(adapters Adapters) error {
	defer func() {
		m.feeds = make(feedsMap)
	}()

	socialGraph, err := adapters.SocialGraph.GetSocialGraph()
	if err != nil {
		return errors.Wrap(err, "could not load the social graph")
	}

	for _, msgs := range m.feeds {
		firstMessage := msgs[0]

		if !socialGraph.HasContact(firstMessage.Author()) {
			continue // do nothing as this contact is not in our social graph
		}

		sort.Slice(msgs, func(i, j int) bool {
			return msgs[j].Sequence().ComesAfter(msgs[i].Sequence())
		})

		if err := adapters.Feed.UpdateFeed(firstMessage.Feed(), func(feed *feeds.Feed) (*feeds.Feed, error) {
			for _, msg := range msgs {
				if err := feed.AppendMessage(msg); err != nil {
					return nil, errors.Wrap(err, "could not append a message")
				}
			}
			return feed, nil
		}); err != nil {
			return errors.Wrap(err, "failed to update the feed")
		}
	}

	return nil
}

type feedsMap map[string][]message.Message

func (f feedsMap) AddMessage(msg message.Message) {
	key := msg.Feed().String()
	f[key] = append(f[key], msg)
}

func (f feedsMap) MessageCount() int {
	var result int
	for _, messages := range f {
		result += len(messages)
	}
	return result
}
