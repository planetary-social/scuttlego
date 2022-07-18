package queries

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const (
	numCreateHistoryStreamWorkers = 10
	noNewRequestsToProcessDelay   = 100 * time.Millisecond
	cleanupDelay                  = 500 * time.Millisecond
)

type FeedRepository interface {
	// GetMessages returns messages with a sequence greater or equal to the
	// provided sequence. If sequence is nil then messages starting from the
	// beginning of the feed are returned. Limit specifies the max number of
	// returned messages. If limit is nil then all messages matching the
	// sequence criteria are returned.
	GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) // todo iterator instead of returning a huge array

	// Count returns the number of stored feeds.
	Count() (int, error)
}

type MessageSubscriber interface {
	// SubscribeToNewMessages subscribes to all new messages.
	SubscribeToNewMessages(ctx context.Context) <-chan message.Message
}

type CreateHistoryStreamResponseWriter interface {
	WriteMessage(msg message.Message) error
	CloseWithError(err error) error
}

type CreateHistoryStream struct {
	Id refs.Feed

	// If not set messages starting from the very beginning of the feed will be
	// returned. Otherwise, messages with a sequence greater or equal to the
	// provided one will be returned. See:
	// %ptQutWwkNIIteEn791Ru27DHtOsdnbcEJRgjuxW90Y4=.sha256
	Seq *message.Sequence

	// Number of messages to return, if not set then unlimited.
	Limit *int

	// If true then the channel will stay open and return further messages as
	// they become available. This usually means returning further messages
	// which are replicated from other peers.
	Live bool

	// Used together with live. If true then old messages will be returned
	// before serving live messages as they come in. You most likely always want
	// to set this to true. Setting Live to false and Old to false probably
	// makes no sense but won't cause an error, however no messages will be
	// returned and the channel will be closed immediately.
	Old bool

	// ResponseWriter is used to send messages to the peer. Firstly WriteMessage
	// will be called zero or more times. Then CloseWithError will be called
	// exactly once.
	ResponseWriter CreateHistoryStreamResponseWriter
}

type CreateHistoryStreamHandler struct {
	repository FeedRepository
	subscriber MessageSubscriber

	queue *RequestQueue

	streams     map[string][]*HistoryStream
	streamsLock sync.Mutex

	logger logging.Logger
}

func NewCreateHistoryStreamHandler(
	repository FeedRepository,
	subscriber MessageSubscriber,
	logger logging.Logger,
) *CreateHistoryStreamHandler {
	return &CreateHistoryStreamHandler{
		repository: repository,
		subscriber: subscriber,

		queue:   NewRequestQueue(),
		streams: make(map[string][]*HistoryStream),

		logger: logger.New("create_history_stream_handler"),
	}
}

func (h *CreateHistoryStreamHandler) Handle(ctx context.Context, query CreateHistoryStream) {
	h.queue.Add(NewCreateHistoryStreamToProcess(ctx, query))
}

func (h *CreateHistoryStreamHandler) Run(ctx context.Context) error {
	h.startWorkers(ctx)
	<-ctx.Done()
	return ctx.Err()
}

func (h *CreateHistoryStreamHandler) startWorkers(ctx context.Context) {
	for i := 0; i < numCreateHistoryStreamWorkers; i++ {
		go h.worker(ctx)
	}
	go h.liveWorker(ctx)
	go h.cleanupWorker(ctx)
}

func (h *CreateHistoryStreamHandler) worker(ctx context.Context) {
	for {
		req, hasNextRequest := h.queue.Get()
		if !hasNextRequest {
			select {
			case <-time.After(noNewRequestsToProcessDelay):
				continue
			case <-ctx.Done():
				return
			}
		}

		if err := h.processRequest(req.Ctx(), req.Query()); err != nil {
			if closeErr := req.Query().ResponseWriter.CloseWithError(err); closeErr != nil {
				h.logger.WithError(closeErr).Debug("closing failed")
			}
			h.logger.WithError(err).Error("error processing an incoming request")
		}
	}
}

func (h *CreateHistoryStreamHandler) liveWorker(ctx context.Context) {
	for msg := range h.subscriber.SubscribeToNewMessages(ctx) {
		h.onLiveMessage(msg)
	}
}

func (h *CreateHistoryStreamHandler) cleanupWorker(ctx context.Context) {
	for {
		h.cleanup()

		select {
		case <-time.After(cleanupDelay):
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (h *CreateHistoryStreamHandler) cleanup() {
	h.streamsLock.Lock()
	defer h.streamsLock.Unlock()

	for key := range h.streams {
		for i := range h.streams[key] {
			if h.streams[key][i].IsClosed() {
				if closeErr := h.streams[key][i].CloseWithError(nil); closeErr != nil {
					h.logger.WithError(closeErr).Debug("closing failed")
				}
				h.streams[key] = append(h.streams[key][:i], h.streams[key][i+1:]...)
			}
		}

		if len(h.streams[key]) == 0 {
			delete(h.streams, key)
		}
	}
}

func (h *CreateHistoryStreamHandler) processRequest(ctx context.Context, query CreateHistoryStream) error {
	s := NewHistoryStream(ctx, query)

	if query.Live {
		h.registerStreamForLive(s)
	}

	if query.Old {
		msgs, err := h.repository.GetMessages(query.Id, query.Seq, query.Limit)
		if err != nil {
			return errors.Wrap(err, "could not retrieve messages")
		}

		for _, msg := range msgs {
			if err := s.OnOldMessage(msg); err != nil {
				return errors.Wrap(err, "error handling an old message")
			}
		}
	}

	if query.Live && !s.ReachedLimit() {
		if err := s.SwitchToLiveMode(); err != nil {
			return errors.Wrap(err, "failed to switch to live mode")
		}
	} else {
		if closeErr := query.ResponseWriter.CloseWithError(nil); closeErr != nil {
			h.logger.WithError(closeErr).Debug("closing failed")
		}
	}

	return nil
}

func (h *CreateHistoryStreamHandler) registerStreamForLive(s *HistoryStream) {
	h.streamsLock.Lock()
	defer h.streamsLock.Unlock()

	key := s.Feed().String()
	h.streams[key] = append(h.streams[key], s)
}

func (h *CreateHistoryStreamHandler) onLiveMessage(msg message.Message) {
	h.streamsLock.Lock()
	defer h.streamsLock.Unlock()

	key := h.feedKey(msg.Feed())

	for i := len(h.streams[key]) - 1; i >= 0; i-- {
		stream := h.streams[key][i]
		if err := stream.OnLiveMessage(msg); err != nil {
			if errors.Is(err, ErrLimitReached) {
				if closeErr := stream.CloseWithError(nil); closeErr != nil {
					h.logger.WithError(closeErr).Debug("closing failed")
				}
			} else {
				if closeErr := stream.CloseWithError(err); closeErr != nil {
					h.logger.WithError(closeErr).Debug("closing failed")
				}
			}
			h.streams[key] = append(h.streams[key][:i], h.streams[key][i+1:]...)
		}
	}
}

func (h *CreateHistoryStreamHandler) feedKey(feed refs.Feed) string {
	return feed.String()
}

type HistoryStream struct {
	ctx context.Context
	rw  CreateHistoryStreamResponseWriter

	feed  refs.Feed
	limit *int
	old   bool
	live  bool

	lastSequence *message.Sequence
	sentMessages int

	queue                   []message.Message
	readyToSendLiveMessages bool
	lock                    sync.Mutex // secures queue, readyToSendLiveMessages
}

func NewHistoryStream(ctx context.Context, query CreateHistoryStream) *HistoryStream {
	return &HistoryStream{
		ctx:          ctx,
		rw:           query.ResponseWriter,
		feed:         query.Id,
		limit:        query.Limit,
		old:          query.Old,
		live:         query.Live,
		lastSequence: query.Seq,
	}
}

func (s *HistoryStream) OnOldMessage(msg message.Message) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
	}

	if !s.old {
		return errors.New("old is set to false")
	}

	return s.sendMessage(msg)
}

func (s *HistoryStream) SwitchToLiveMode() error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
	}

	if !s.live {
		return errors.New("live is set to false")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	sort.Slice(s.queue, func(i, j int) bool {
		return !s.queue[i].Sequence().ComesAfter(s.queue[j].Sequence())
	})

	for _, msg := range s.queue {
		if err := s.rw.WriteMessage(msg); err != nil {
			return errors.Wrap(err, "failed to write message")
		}
	}

	s.readyToSendLiveMessages = true
	return nil
}

var ErrLimitReached = errors.New("limit reached")

func (s *HistoryStream) OnLiveMessage(msg message.Message) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
	}

	if !s.live {
		return errors.New("live is set to false")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.readyToSendLiveMessages {
		s.queue = append(s.queue, msg)
		return nil
	}

	if s.limit != nil && s.sentMessages >= *s.limit {
		return ErrLimitReached
	}

	if !msg.Feed().Equal(s.feed) {
		return errors.New("message doesn't belong to this feed")
	}

	if s.lastSequence != nil && !msg.Sequence().ComesAfter(*s.lastSequence) {
		return nil
	}

	return s.sendMessage(msg)
}

func (s *HistoryStream) CloseWithError(err error) error {
	return s.rw.CloseWithError(err)
}

func (s *HistoryStream) sendMessage(msg message.Message) error {
	if err := s.rw.WriteMessage(msg); err != nil {
		return errors.Wrap(err, "failed to write message")
	}

	seq := msg.Sequence()
	s.lastSequence = &seq
	s.sentMessages++
	return nil
}

func (s *HistoryStream) Feed() refs.Feed {
	return s.feed
}

func (s *HistoryStream) ReachedLimit() bool {
	return s.limit != nil && s.sentMessages >= *s.limit
}

func (s *HistoryStream) IsClosed() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}

type RequestQueue struct {
	queue     []CreateHistoryStreamToProcess
	queueLock sync.Mutex
}

func NewRequestQueue() *RequestQueue {
	return &RequestQueue{}
}

func (q *RequestQueue) Add(v CreateHistoryStreamToProcess) {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()

	q.queue = append(q.queue, v)
}

func (q *RequestQueue) Get() (CreateHistoryStreamToProcess, bool) {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()

	if len(q.queue) == 0 {
		return CreateHistoryStreamToProcess{}, false
	}

	req := q.queue[0]
	q.queue = q.queue[1:]
	return req, true
}

type CreateHistoryStreamToProcess struct {
	ctx   context.Context
	query CreateHistoryStream
}

func NewCreateHistoryStreamToProcess(ctx context.Context, query CreateHistoryStream) CreateHistoryStreamToProcess {
	return CreateHistoryStreamToProcess{ctx: ctx, query: query}
}

func (c CreateHistoryStreamToProcess) Ctx() context.Context {
	return c.ctx
}

func (c CreateHistoryStreamToProcess) Query() CreateHistoryStream {
	return c.query
}
