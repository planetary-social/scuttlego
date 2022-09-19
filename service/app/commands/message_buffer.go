package commands

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const (
	messageBufferPersistAtMessages = 1000
	messageBufferPersistEvery      = 5 * time.Second
)

type MessageBuffer struct {
	feeds     feedsMap
	feedsLock *sync.Mutex

	forcePersistCh   chan struct{}
	forcePersistOnce sync.Once

	transaction TransactionProvider
	logger      logging.Logger
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

	m.feeds.AddMessage(msg)

	if m.feeds.MessageCount() > messageBufferPersistAtMessages {
		m.forcePersistOnce.Do(
			func() {
				close(m.forcePersistCh)
			},
		)
	}

	return nil
}

func (m *MessageBuffer) Sequence(feed refs.Feed) (message.Sequence, bool) {
	m.feedsLock.Lock()
	defer m.feedsLock.Unlock()

	msgs := m.feeds[feed.String()]
	if len(msgs) == 0 {
		return message.Sequence{}, false
	}

	return msgs[len(msgs)-1].Sequence(), true
}

func (m *MessageBuffer) persist() error {
	m.feedsLock.Lock()
	defer m.feedsLock.Unlock()

	defer func() {
		m.feeds = make(feedsMap)
		m.forcePersistCh = make(chan struct{})
		m.forcePersistOnce = sync.Once{}
	}()

	numberOfMessages := m.feeds.MessageCount()
	if numberOfMessages == 0 {
		return nil
	}

	start := time.Now()

	if err := m.transaction.Transact(func(adapters Adapters) error {
		return m.persistTransaction(adapters, m.feeds)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	m.logger.
		WithField("count", numberOfMessages).
		WithField("duration", time.Since(start)).
		Debug("persisted messages")

	return nil
}

func (m *MessageBuffer) persistTransaction(adapters Adapters, feedsToPersist feedsMap) error {
	socialGraph, err := adapters.SocialGraph.GetSocialGraph()
	if err != nil {
		return errors.Wrap(err, "could not load the social graph")
	}

	for _, msgs := range feedsToPersist {
		firstMessage := msgs[0]

		authorRef := firstMessage.Author()
		feedRef := firstMessage.Feed()

		if !socialGraph.HasContact(authorRef) {
			continue // do nothing as this contact is not in our social graph
		}

		if err := adapters.Feed.UpdateFeed(feedRef, func(feed *feeds.Feed) error {
			for _, msg := range msgs {
				if err := feed.AppendMessage(msg); err != nil {
					return errors.Wrap(err, "could not append a message")
				}
			}
			return nil
		}); err != nil {
			return errors.Wrapf(err, "failed to update the feed '%s'", feedRef)
		}
	}

	return nil
}

type feedsMap map[string][]message.Message

func (f feedsMap) AddMessage(msg message.Message) {
	key := msg.Feed().String()
	f[key] = append(f[key], msg)

	if msgs := f[key]; len(msgs) > 1 {
		last := len(msgs) - 1
		if !msgs[last].Sequence().ComesAfter(msgs[last-1].Sequence()) {
			sort.Slice(msgs, func(i, j int) bool {
				return msgs[j].Sequence().ComesAfter(msgs[i].Sequence())
			})
		}
	}
}

func (f feedsMap) MessageCount() int {
	var result int
	for _, messages := range f {
		result += len(messages)
	}
	return result
}
