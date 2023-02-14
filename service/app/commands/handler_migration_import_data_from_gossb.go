package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	gossbrefs "github.com/ssbc/go-ssb-refs"
)

const (
	maxMessagesInMemory         = 1000
	convertedMessagesBufferSize = 1000
	maxMessagesPerTransaction   = 50
	minSequenceProgressToSave   = 1000
)

type GoSSBRepoReader interface {
	// GetMessages returns a channel on which receive log messages will be sent
	// ordered by their receive log sequence. Some messages may be missing. If
	// the provided value is not nil it will resume from the provided sequence.
	GetMessages(ctx context.Context, directory string, resumeFromSequence *common.ReceiveLogSequence) (<-chan GoSSBMessageOrError, error)
}

type ContentParser interface {
	Parse(raw message.RawMessageContent) (message.Content, error)
}

type GoSSBMessageOrError struct {
	Value GoSSBMessage
	Err   error
}

type GoSSBMessage struct {
	ReceiveLogSequence common.ReceiveLogSequence
	Message            gossbrefs.Message
}

type SaveResumeFromSequenceFn func(common.ReceiveLogSequence) error

type ImportDataFromGoSSB struct {
	directory                string
	resumeFromSequence       *common.ReceiveLogSequence
	saveResumeFromSequenceFn SaveResumeFromSequenceFn
}

func NewImportDataFromGoSSB(
	directory string,
	resumeFromSequence *common.ReceiveLogSequence,
	saveResumeFromSequenceFn SaveResumeFromSequenceFn,
) (ImportDataFromGoSSB, error) {
	if directory == "" {
		return ImportDataFromGoSSB{}, errors.New("directory is an empty string")
	}
	if saveResumeFromSequenceFn == nil {
		return ImportDataFromGoSSB{}, errors.New("nil save resume from sequence function")
	}
	return ImportDataFromGoSSB{
		directory:                directory,
		resumeFromSequence:       resumeFromSequence,
		saveResumeFromSequenceFn: saveResumeFromSequenceFn,
	}, nil
}

func (cmd ImportDataFromGoSSB) IsZero() bool {
	return cmd.directory == ""
}

type ImportDataFromGoSSBResult struct {
	Successes int
	Errors    int
}

type MigrationHandlerImportDataFromGoSSB struct {
	repoReader    GoSSBRepoReader
	transaction   TransactionProvider
	contentParser ContentParser
	logger        logging.Logger
}

func NewMigrationHandlerImportDataFromGoSSB(
	repoReader GoSSBRepoReader,
	transaction TransactionProvider,
	contentParser ContentParser,
	logger logging.Logger,
) *MigrationHandlerImportDataFromGoSSB {
	return &MigrationHandlerImportDataFromGoSSB{
		repoReader:    repoReader,
		transaction:   transaction,
		contentParser: contentParser,
		logger:        logger.New("migration_handler_import_data_from_go_ssb"),
	}
}

func (h MigrationHandlerImportDataFromGoSSB) Handle(ctx context.Context, cmd ImportDataFromGoSSB) (ImportDataFromGoSSBResult, error) {
	if cmd.IsZero() {
		return ImportDataFromGoSSBResult{}, errors.New("zero value of command")
	}

	h.logger.
		WithField("directory", cmd.directory).
		WithField("resume_from_sequence", cmd.resumeFromSequence).
		Debug("import starting")

	msgs := newMessagesToImport()
	successCounter := 0
	errorCounter := 0
	start := time.Now()

	defer func() {
		h.logger.
			WithField("success_counter", successCounter).
			WithField("error_counter", errorCounter).
			WithField("elapsed_time", time.Since(start).String()).
			Debug("import ended")
	}()

	gossbMessageCh, err := h.repoReader.GetMessages(ctx, cmd.directory, cmd.resumeFromSequence)
	if err != nil {
		return ImportDataFromGoSSBResult{}, errors.Wrap(err, "error getting message channel")
	}

	lastSequenceSaved := cmd.resumeFromSequence

	for v := range h.startConversionLoop(ctx, gossbMessageCh) {
		if err := v.Err; err != nil {
			return ImportDataFromGoSSBResult{}, errors.Wrap(err, "received an error")
		}

		if err := msgs.AddMessage(v.Value); err != nil {
			return ImportDataFromGoSSBResult{}, errors.Wrap(err, "error adding message to buffer")
		}

		if msgs.MessageCount() >= maxMessagesInMemory {
			if err := h.saveAllMessages(
				msgs,
				&errorCounter,
				&successCounter,
				cmd.saveResumeFromSequenceFn,
				&lastSequenceSaved,
			); err != nil {
				return ImportDataFromGoSSBResult{}, errors.Wrap(err, "error saving messages")
			}
		} else {
			if perFeedMsgs := msgs.Get(v.Value.Message.Feed()); perFeedMsgs.Len() >= maxMessagesPerTransaction {
				if err := h.saveMessagesPerFeed(
					msgs,
					perFeedMsgs.Feed(),
					&errorCounter,
					&successCounter,
					cmd.saveResumeFromSequenceFn,
					&lastSequenceSaved,
				); err != nil {
					return ImportDataFromGoSSBResult{}, errors.Wrap(err, "error saving messages per feed")
				}
			}
		}
	}

	if err := h.saveAllMessages(
		msgs,
		&errorCounter,
		&successCounter,
		cmd.saveResumeFromSequenceFn,
		&lastSequenceSaved,
	); err != nil {
		return ImportDataFromGoSSBResult{}, errors.Wrap(err, "error saving messages")
	}

	return ImportDataFromGoSSBResult{
		Successes: successCounter,
		Errors:    errorCounter,
	}, nil
}

func (h MigrationHandlerImportDataFromGoSSB) startConversionLoop(ctx context.Context, in <-chan GoSSBMessageOrError) <-chan scuttlegoMessageOrError {
	out := make(chan scuttlegoMessageOrError, convertedMessagesBufferSize)

	go func() {
		defer close(out)

		for v := range in {
			if err := v.Err; err != nil {
				select {
				case out <- scuttlegoMessageOrError{Err: errors.Wrap(err, "received an error")}:
					return
				case <-ctx.Done():
					return
				}
			}

			h.logger.
				WithField("receive_log_sequence", v.Value.ReceiveLogSequence.Int()).
				WithField("message_id", v.Value.Message.Key().Sigil()).
				Trace("converting message")

			msg, err := h.convertMessage(v.Value.Message)
			if err != nil {
				select {
				case out <- scuttlegoMessageOrError{Err: errors.Wrap(err, "convert message error")}:
					return
				case <-ctx.Done():
					return
				}
			}

			value := scuttlegoMessage{
				ReceiveLogSequence: v.Value.ReceiveLogSequence,
				Message:            msg,
			}

			select {
			case out <- scuttlegoMessageOrError{Value: value}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

func (h MigrationHandlerImportDataFromGoSSB) saveAllMessages(
	values messagesToImport,
	errorCounter *int,
	successCounter *int,
	saveResumeFromSequenceFn SaveResumeFromSequenceFn,
	lastSavedSequencePtr **common.ReceiveLogSequence,
) error {
	for _, perFeedMsgs := range values {
		if err := h.saveMessagesPerFeed(values, perFeedMsgs.Feed(), errorCounter, successCounter, saveResumeFromSequenceFn, lastSavedSequencePtr); err != nil {
			return errors.Wrap(err, "error saving messages per feed")
		}
	}

	return nil
}

func (h MigrationHandlerImportDataFromGoSSB) saveMessagesPerFeed(
	msgs messagesToImport,
	feed refs.Feed,
	errorCounter *int,
	successCounter *int,
	saveResumeFromSequenceFn SaveResumeFromSequenceFn,
	lastSequenceSavedPtr **common.ReceiveLogSequence,
) error {
	msgsPerFeed := msgs.Get(feed)

	if msgsPerFeed.Len() == 0 {
		return nil
	}

	start := time.Now()
	var feedErrors, feedSuccesses int

	if err := h.transaction.Transact(func(adapters Adapters) error {
		feedErrors = 0
		feedSuccesses = 0

		var successes []scuttlegoMessage

		if err := adapters.Feed.UpdateFeedIgnoringReceiveLog(feed, func(feed *feeds.Feed) error {
			for _, msg := range msgsPerFeed.messages {
				if err := feed.AppendMessage(msg.Message); err != nil {
					h.logger.
						WithError(err).
						WithField("msg.feed", msg.Message.Feed().String()).
						WithField("msg.id", msg.Message.Id().String()).
						WithField("msg.sequence", msg.Message.Sequence().Int()).
						WithField("receive_log_sequence", msg.ReceiveLogSequence.Int()).
						Error("error appending a message")
					feedErrors += 1
					continue
				}

				feedSuccesses += 1
				successes = append(successes, msg)
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "error updating the feed")
		}

		for _, msg := range successes {
			foundMessage, err := adapters.ReceiveLog.GetMessage(msg.ReceiveLogSequence)
			if err != nil {
				if !errors.Is(err, common.ErrReceiveLogEntryNotFound) {
					return errors.Wrap(err, "error getting message from receive log")
				}
			} else {
				if !foundMessage.Id().Equal(msg.Message.Id()) {
					return fmt.Errorf(
						"duplicate message with receive log sequence '%d', old='%s', new='%s'",
						msg.ReceiveLogSequence.Int(),
						foundMessage.Id(),
						msg.Message.Id(),
					)
				}
			}

			if err := adapters.ReceiveLog.PutUnderSpecificSequence(msg.Message.Id(), msg.ReceiveLogSequence); err != nil {
				return errors.Wrap(err, "error updating the receive log")
			}
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	h.logger.
		WithField("successes", feedSuccesses).
		WithField("errors", feedErrors).
		WithField("elapsed", time.Since(start).String()).
		Trace("saved messages for feed")

	*errorCounter += feedErrors
	*successCounter += feedSuccesses

	if err := h.maybePersistSequence(msgs, saveResumeFromSequenceFn, lastSequenceSavedPtr); err != nil {
		return errors.Wrap(err, "error saving messages")
	}

	msgs.Delete(feed)

	return nil
}

func (h MigrationHandlerImportDataFromGoSSB) maybePersistSequence(
	msgs messagesToImport,
	fn SaveResumeFromSequenceFn,
	lastSavedSequencePtr **common.ReceiveLogSequence,
) error {
	sequenceToSave, err := h.lowestReceiveLogSequence(msgs)
	if err != nil {
		return errors.Wrap(err, "error determining sequence to persist")
	}

	if lastSavedSequence := *lastSavedSequencePtr; lastSavedSequence != nil && sequenceToSave.Int() < lastSavedSequence.Int()+minSequenceProgressToSave {
		return nil
	}

	if err := fn(sequenceToSave); err != nil {
		return errors.Wrap(err, "error saving the receive log sequence")
	}

	h.logger.WithField("sequence", sequenceToSave.Int()).Debug("saved the receive log sequence")
	*lastSavedSequencePtr = &sequenceToSave

	return nil
}

func (h MigrationHandlerImportDataFromGoSSB) lowestReceiveLogSequence(msgs messagesToImport) (common.ReceiveLogSequence, error) {
	var lowest *common.ReceiveLogSequence

	for _, v := range msgs {
		if lowest == nil || lowest.Int() > v.messages[0].ReceiveLogSequence.Int() {
			tmp := v.messages[0].ReceiveLogSequence
			lowest = &tmp
		}
	}

	if lowest == nil {
		return common.ReceiveLogSequence{}, errors.New("empty messages")
	}

	return *lowest, nil
}

func (h MigrationHandlerImportDataFromGoSSB) convertMessage(gossbmsg gossbrefs.Message) (message.Message, error) {
	rawMessage, err := message.NewRawMessage(gossbmsg.ValueContentJSON())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error creating a raw message")
	}

	id, err := refs.NewMessage(gossbmsg.Key().Sigil())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error creating id")
	}

	var previous *refs.Message
	if gossbmsg.Previous() != nil {
		tmp, err := refs.NewMessage(gossbmsg.Previous().Sigil())
		if err != nil {
			return message.Message{}, errors.Wrap(err, "error creating previous message id")
		}
		previous = &tmp
	}

	sequence, err := message.NewSequence(int(gossbmsg.Seq()))
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error creating sequence")
	}

	author, err := refs.NewIdentity(gossbmsg.Author().Sigil())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error creating author")
	}

	feed, err := refs.NewFeed(gossbmsg.Author().Sigil())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error creating feed")
	}

	rawMessageContent, err := message.NewRawMessageContent(gossbmsg.ContentBytes())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error creating raw message content")
	}

	messageContent, err := h.contentParser.Parse(rawMessageContent)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error parsing content")
	}

	msg, err := message.NewMessage(
		id,
		previous,
		sequence,
		author,
		feed,
		gossbmsg.Claimed(),
		messageContent,
		rawMessage,
	)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error creating a message")
	}

	return msg, nil
}

type scuttlegoMessage struct {
	ReceiveLogSequence common.ReceiveLogSequence
	Message            message.Message
}

type scuttlegoMessageOrError struct {
	Value scuttlegoMessage
	Err   error
}

type messagesToImport map[string]*messagesToImportPerFeed

func newMessagesToImport() messagesToImport {
	return make(messagesToImport)
}

func (f messagesToImport) AddMessage(msg scuttlegoMessage) error {
	return f.Get(msg.Message.Feed()).Add(msg)
}

func (f messagesToImport) Get(feed refs.Feed) *messagesToImportPerFeed {
	key := feed.String()
	v, ok := f[key]
	if !ok {
		v = newMessagesToImportPerFeed(feed)
		f[key] = v
	}
	return v
}

func (f messagesToImport) Delete(feed refs.Feed) {
	delete(f, feed.String())
}

func (f messagesToImport) MessageCount() int {
	var result int
	for _, messages := range f {
		result += messages.Len()
	}
	return result
}

type messagesToImportPerFeed struct {
	feed     refs.Feed
	messages []scuttlegoMessage
}

func newMessagesToImportPerFeed(feed refs.Feed) *messagesToImportPerFeed {
	return &messagesToImportPerFeed{feed: feed}
}

func (m *messagesToImportPerFeed) Add(msg scuttlegoMessage) error {
	if !msg.Message.Feed().Equal(m.feed) {
		return errors.New("incorrect feed")
	}

	m.messages = append(m.messages, msg)
	return nil
}

func (m *messagesToImportPerFeed) Len() int {
	return len(m.messages)
}

func (m *messagesToImportPerFeed) Feed() refs.Feed {
	return m.feed
}
