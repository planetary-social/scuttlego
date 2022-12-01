package commands

import (
	"context"

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
	GetMessages(ctx context.Context, directory string) (<-chan GoSSBMessageOrError, error)
}

type GoSSBMessageOrError struct {
	Value GoSSBMessage
	Err   error
}

type GoSSBMessage struct {
	ReceiveLogSequence int64
	Message            gossbrefs.Message
}

type ImportDataFromGoSSB struct {
	directory string
}

func NewImportDataFromGoSSB(directory string) (ImportDataFromGoSSB, error) {
	if directory == "" {
		return ImportDataFromGoSSB{}, errors.New("directory is an empty string")
	}
	return ImportDataFromGoSSB{directory: directory}, nil
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
	ch, err := h.repoReader.GetMessages(ctx, cmd.directory)
	if err != nil {
		return errors.Wrap(err, "error getting message channel")
	}

	var msgs []GoSSBMessage

	for v := range ch {
		if err := v.Err; err != nil {
			return errors.Wrap(err, "received an error")
		}

		h.logger.
			WithField("receive_log_sequence", v.Value.ReceiveLogSequence).
			WithField("message_id", v.Value.Message.Key().Sigil()).
			Trace("processing message")

		msgs = append(msgs, v.Value)

		if len(msgs) >= batchSize {
			if err := h.saveMessages(msgs); err != nil {
				return errors.Wrap(err, "error saving messages")
			}
			msgs = nil
		}
	}

	if err := h.saveMessages(msgs); err != nil {
		return errors.Wrap(err, "error saving messages")
	}

	return nil
}

func (h MigrationHandlerImportDataFromGoSSB) saveMessages(gossbmsgs []GoSSBMessage) error {
	if err := h.transaction.Transact(func(adapters Adapters) error {
		for _, gossbmsg := range gossbmsgs {
			if err := h.saveMessage(adapters, gossbmsg); err != nil {
				return errors.Wrapf(err, "error saving message '%s' with receive log sequence '%d'", gossbmsg.Message.Key().Sigil(), gossbmsg.ReceiveLogSequence)
			}
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func (h MigrationHandlerImportDataFromGoSSB) saveMessage(adapters Adapters, gossbmsg GoSSBMessage) error {
	msg, err := h.convertMessage(gossbmsg.Message)
	if err != nil {
		return errors.Wrap(err, "error converting the message")
	}

	receiveLogSequence, err := common.NewReceiveLogSequence(int(gossbmsg.ReceiveLogSequence))
	if err != nil {
		return errors.Wrap(err, "error creating the receive log sequence")
	}

	if err := adapters.Feed.UpdateFeedIgnoringReceiveLog(msg.Feed(), func(feed *feeds.Feed) error {
		return feed.AppendMessage(msg)
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
