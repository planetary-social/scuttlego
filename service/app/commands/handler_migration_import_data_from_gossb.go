package commands

import (
	"context"
	"fmt"

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

func (h MigrationHandlerImportDataFromGoSSB) Handle(ctx context.Context, cmd ImportDataFromGoSSB) error {
	ch, err := h.repoReader.GetMessages(ctx, cmd.directory, cmd.resumeAfterSequence)
	if err != nil {
		return errors.Wrap(err, "error getting message channel")
	}

	var msgs []GoSSBMessage

	for v := range ch {
		if err := v.Err; err != nil {
			return errors.Wrap(err, "received an error")
		}

		h.logger.
			WithField("receive_log_sequence", v.Value.ReceiveLogSequence.Int()).
			WithField("message_id", v.Value.Message.Key().Sigil()).
			Trace("processing message")

		msgs = append(msgs, v.Value)

		if len(msgs) >= batchSize {
			if err := h.saveMessages(msgs, cmd.saveResumeAfterSequenceFn); err != nil {
				return errors.Wrap(err, "error saving messages")
			}
			msgs = nil
		}
	}

	if err := h.saveMessages(msgs, cmd.saveResumeAfterSequenceFn); err != nil {
		return errors.Wrap(err, "error saving messages")
	}

	return nil
}

func (h MigrationHandlerImportDataFromGoSSB) saveMessages(
	gossbmsgs []GoSSBMessage,
	saveResumeAfterSequenceFn SaveResumeAfterSequenceFn,
) error {
	if len(gossbmsgs) == 0 {
		return nil
	}

	if err := h.transaction.Transact(func(adapters Adapters) error {
		for _, gossbmsg := range gossbmsgs {
			if err := h.saveMessage(adapters, gossbmsg); err != nil {
				return errors.Wrapf(err, "error saving message '%s' with receive log sequence '%d'", gossbmsg.Message.Key().Sigil(), gossbmsg.ReceiveLogSequence.Int())
			}
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	lastMsgReceiveLogSequence := gossbmsgs[len(gossbmsgs)-1].ReceiveLogSequence
	if err := saveResumeAfterSequenceFn(lastMsgReceiveLogSequence); err != nil {
		return errors.Wrap(err, "error saving the receive log sequence")
	}

	return nil
}

func (h MigrationHandlerImportDataFromGoSSB) saveMessage(adapters Adapters, gossbmsg GoSSBMessage) error {
	msg, err := h.convertMessage(gossbmsg.Message)
	if err != nil {
		return errors.Wrap(err, "error converting the message")
	}

	foundMessage, err := adapters.ReceiveLog.GetMessage(gossbmsg.ReceiveLogSequence)
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
		return feed.AppendMessage(msg)
	}); err != nil {
		return errors.Wrap(err, "error updating the feed")
	}

	if err := adapters.ReceiveLog.PutUnderSpecificSequence(msg.Id(), gossbmsg.ReceiveLogSequence); err != nil {
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
