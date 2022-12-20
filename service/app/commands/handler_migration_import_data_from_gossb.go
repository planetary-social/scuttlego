package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	gossbrefs "go.mindeco.de/ssb-refs"
)

const batchSize = 5000

type GoSSBRepoReader interface {
	// GetMessages returns a channel on which receive log messages will be sent
	// ordered by their receive log sequence. Some messages may be missing. If
	// the provided value is not nil it will resume after the provided sequence.
	GetMessages(ctx context.Context, directory string, resumeAfterSequence *common.ReceiveLogSequence) (<-chan GoSSBMessageOrError, error)
}

type GoSSBMessageOrError struct {
	Value GoSSBMessage
	Err   error
}

type GoSSBMessage struct {
	ReceiveLogSequence common.ReceiveLogSequence
	Message            gossbrefs.Message
}

type SaveResumeAfterSequenceFn func(common.ReceiveLogSequence) error

type ImportDataFromGoSSB struct {
	directory                 string
	resumeAfterSequence       *common.ReceiveLogSequence
	saveResumeAfterSequenceFn SaveResumeAfterSequenceFn
}

func NewImportDataFromGoSSB(
	directory string,
	resumeAfterSequence *common.ReceiveLogSequence,
	saveResumeAfterSequenceFn SaveResumeAfterSequenceFn,
) (ImportDataFromGoSSB, error) {
	if directory == "" {
		return ImportDataFromGoSSB{}, errors.New("directory is an empty string")
	}
	if saveResumeAfterSequenceFn == nil {
		return ImportDataFromGoSSB{}, errors.New("nil save resume after sequence function")
	}
	return ImportDataFromGoSSB{
		directory:                 directory,
		resumeAfterSequence:       resumeAfterSequence,
		saveResumeAfterSequenceFn: saveResumeAfterSequenceFn,
	}, nil
}

type ImportDataFromGoSSBResult struct {
	Successes int
	Errors    int
}

type MigrationHandlerImportDataFromGoSSB struct {
	repoReader  GoSSBRepoReader
	transaction TransactionProvider
	marshaler   formats.Marshaler
	logger      logging.Logger
}

func NewMigrationHandlerImportDataFromGoSSB(
	repoReader GoSSBRepoReader,
	transaction TransactionProvider,
	marshaler formats.Marshaler,
	logger logging.Logger,
) *MigrationHandlerImportDataFromGoSSB {
	return &MigrationHandlerImportDataFromGoSSB{
		repoReader:  repoReader,
		transaction: transaction,
		marshaler:   marshaler,
		logger:      logger.New("migration_handler_import_data_from_go_ssb"),
	}
}

func (h MigrationHandlerImportDataFromGoSSB) Handle(ctx context.Context, cmd ImportDataFromGoSSB) (ImportDataFromGoSSBResult, error) {
	h.logger.
		WithField("directory", cmd.directory).
		WithField("resume_after_sequence", cmd.resumeAfterSequence).
		Debug("import starting")

	msgs := make([]scuttlegoMessage, 0, batchSize)
	successCounter := 0
	errorCounter := 0
	start := time.Now()

	defer func() {
		h.logger.
			WithField("success_counter", successCounter).
			WithField("error_counter", errorCounter).
			WithField("elapsed_time_in_seconds", time.Since(start).String()).
			Debug("import ended")
	}()

	gossbMessageCh, err := h.repoReader.GetMessages(ctx, cmd.directory, cmd.resumeAfterSequence)
	if err != nil {
		return ImportDataFromGoSSBResult{}, errors.Wrap(err, "error getting message channel")
	}

	scuttlegoMessageCh := h.startConversionLoop(ctx, gossbMessageCh)

	for v := range scuttlegoMessageCh {
		if err := v.Err; err != nil {
			return ImportDataFromGoSSBResult{}, errors.Wrap(err, "received an error")
		}

		msgs = append(msgs, v.Value)

		if len(msgs) >= batchSize {
			if err := h.saveMessages(msgs, cmd.saveResumeAfterSequenceFn, &errorCounter, &successCounter); err != nil {
				return ImportDataFromGoSSBResult{}, errors.Wrap(err, "error saving messages")
			}
			msgs = msgs[:0]
		}
	}

	if err := h.saveMessages(msgs, cmd.saveResumeAfterSequenceFn, &errorCounter, &successCounter); err != nil {
		return ImportDataFromGoSSBResult{}, errors.Wrap(err, "error saving messages")
	}

	return ImportDataFromGoSSBResult{
		Successes: successCounter,
		Errors:    errorCounter,
	}, nil
}

func (h MigrationHandlerImportDataFromGoSSB) startConversionLoop(ctx context.Context, in <-chan GoSSBMessageOrError) <-chan scuttlegoMessageOrError {
	out := make(chan scuttlegoMessageOrError, batchSize)

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
				if err := v.Err; err != nil {
					select {
					case out <- scuttlegoMessageOrError{Err: errors.Wrap(err, "convert message error")}:
						return
					case <-ctx.Done():
						return
					}
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

func (h MigrationHandlerImportDataFromGoSSB) saveMessages(
	values []scuttlegoMessage,
	saveResumeAfterSequenceFn SaveResumeAfterSequenceFn,
	errorCounter *int,
	successCounter *int,
) error {
	if len(values) == 0 {
		return nil
	}

	tmpErrorCounter := 0
	tmpSuccessCounter := 0

	start := time.Now()

	if err := h.transaction.Transact(func(adapters Adapters) error {
		tmpErrorCounter = 0
		tmpSuccessCounter = 0

		for _, value := range values {
			if err := h.saveMessage(adapters, value.Message, value.ReceiveLogSequence); err != nil {
				if errors.Is(err, appendMessageError{}) {
					tmpErrorCounter++
					continue
				}
				return errors.Wrapf(err, "error saving message '%s' with receive log sequence '%d'", value.Message.Id().String(), value.ReceiveLogSequence.Int())
			} else {
				tmpSuccessCounter++
			}
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	*errorCounter += tmpErrorCounter
	*successCounter += tmpSuccessCounter

	lastMsgReceiveLogSequence := values[len(values)-1].ReceiveLogSequence
	if err := saveResumeAfterSequenceFn(lastMsgReceiveLogSequence); err != nil {
		return errors.Wrap(err, "error saving the receive log sequence")
	}

	h.logger.
		WithField("successes", tmpSuccessCounter).
		WithField("errors", tmpErrorCounter).
		WithField("elapsed", time.Since(start).String()).
		WithField("last_message_receive_log_sequence", lastMsgReceiveLogSequence.Int()).
		Debug("saved messages")

	return nil
}

func (h MigrationHandlerImportDataFromGoSSB) saveMessage(
	adapters Adapters,
	msg message.Message,
	receiveLogSequence common.ReceiveLogSequence,
) error {
	foundMessage, err := adapters.ReceiveLog.GetMessage(receiveLogSequence)
	if err != nil {
		if !errors.Is(err, common.ErrReceiveLogEntryNotFound) {
			return errors.Wrap(err, "error getting message from receive log")
		}
	} else {
		if !foundMessage.Id().Equal(msg.Id()) {
			return fmt.Errorf("duplicate message, old='%s', new='%s'", foundMessage.Id(), msg.Id())
		}
	}

	if err := adapters.Feed.UpdateFeedIgnoringReceiveLog(msg.Feed(), func(feed *feeds.Feed) error {
		if err := feed.AppendMessage(msg); err != nil {
			return newAppendMessageError(err)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "error updating the feed")
	}

	if err := adapters.ReceiveLog.PutUnderSpecificSequence(msg.Id(), receiveLogSequence); err != nil {
		return errors.Wrap(err, "error updating the receive log")
	}

	return nil
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

	content, err := h.marshaler.Unmarshal(rawMessageContent)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error unmarshaling content")
	}

	msg, err := message.NewMessage(
		id,
		previous,
		sequence,
		author,
		feed,
		gossbmsg.Claimed(),
		content,
		rawMessage,
	)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error creating a message")
	}

	return msg, nil
}

type appendMessageError struct {
	error error
}

func newAppendMessageError(error error) error {
	return &appendMessageError{error: error}
}

func (e appendMessageError) Error() string {
	return "error appending a message"
}

func (e appendMessageError) Unwrap() error {
	return e.error
}

func (e appendMessageError) As(target interface{}) bool {
	if v, ok := target.(*appendMessageError); ok {
		*v = e
		return true
	}
	return false
}

func (e appendMessageError) Is(target error) bool {
	_, ok1 := target.(*appendMessageError)
	_, ok2 := target.(appendMessageError)
	return ok1 || ok2
}

type scuttlegoMessage struct {
	ReceiveLogSequence common.ReceiveLogSequence
	Message            message.Message
}

type scuttlegoMessageOrError struct {
	Value scuttlegoMessage
	Err   error
}
