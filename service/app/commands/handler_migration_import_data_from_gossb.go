package commands

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	gossbrefs "go.mindeco.de/ssb-refs"
)

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
	identifier  RawMessageIdentifier
	logger      logging.Logger
}

func NewMigrationHandlerImportDataFromGoSSB(
	repoReader GoSSBRepoReader,
	transaction TransactionProvider,
	identifier RawMessageIdentifier,
	logger logging.Logger,
) *MigrationHandlerImportDataFromGoSSB {
	return &MigrationHandlerImportDataFromGoSSB{
		repoReader:  repoReader,
		transaction: transaction,
		identifier:  identifier,
		logger:      logger.New("migration_handler_import_data_from_go_ssb"),
	}
}

func (h MigrationHandlerImportDataFromGoSSB) Handle(ctx context.Context, cmd ImportDataFromGoSSB) error {
	ch, err := h.repoReader.GetMessages(ctx, cmd.directory)
	if err != nil {
		return errors.Wrap(err, "error getting message channel")
	}

	for v := range ch {
		if err := v.Err; err != nil {
			return errors.Wrap(err, "received an error")
		}

		if v.Value.ReceiveLogSequence%1000 == 0 {
			fmt.Println(v.Value.ReceiveLogSequence)
		}

		msg, err := h.convertMessage(v.Value.Message)
		if err != nil {
			return errors.Wrap(err, "error converting the message")
		}

		receiveLogSequence, err := common.NewReceiveLogSequence(int(v.Value.ReceiveLogSequence))
		if err != nil {
			return errors.Wrap(err, "error creating the receive log sequence")
		}

		h.logger.
			WithField("receive_log_sequence", receiveLogSequence.Int()).
			WithField("message_id", msg.Id()).
			Debug("processing message")

		if err := h.transaction.Transact(func(adapters Adapters) error {
			if err := adapters.Feed.UpdateFeedIgnoringReceiveLog(msg.Feed(), func(feed *feeds.Feed) error {
				return feed.AppendMessage(msg)
			}); err != nil {
				return errors.Wrap(err, "error updating the feed")
			}

			if err := adapters.ReceiveLog.PutUnderSpecificSequence(msg.Id(), receiveLogSequence); err != nil {
				return errors.Wrap(err, "error updating the feed")
			}

			return nil
		}); err != nil {
			return errors.Wrap(err, "transaction failed")
		}
	}

	return nil
}

func (h MigrationHandlerImportDataFromGoSSB) convertMessage(gossbmsg gossbrefs.Message) (message.Message, error) {
	rawMessage, err := message.NewRawMessage(gossbmsg.ValueContentJSON())
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error creating a raw message")
	}

	msg, err := h.identifier.IdentifyRawMessage(rawMessage)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "error identifying the raw message")
	}

	return msg, nil
}
