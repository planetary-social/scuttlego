package commands

import (
	"context"
	"os"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
)

type DeleteGoSSBRepositoryInOldFormat struct {
	directory string
}

func NewDeleteGoSSBRepositoryInOldFormat(
	directory string,
) (DeleteGoSSBRepositoryInOldFormat, error) {
	if directory == "" {
		return DeleteGoSSBRepositoryInOldFormat{}, errors.New("directory is an empty string")
	}
	return DeleteGoSSBRepositoryInOldFormat{
		directory: directory,
	}, nil
}

func (cmd DeleteGoSSBRepositoryInOldFormat) IsZero() bool {
	return cmd == DeleteGoSSBRepositoryInOldFormat{}
}

type MigrationHandlerDeleteGoSSBRepositoryInOldFormat struct {
	repoReader GoSSBRepoReader
	logger     logging.Logger
}

func NewMigrationHandlerDeleteGoSSBRepositoryInOldFormat(
	repoReader GoSSBRepoReader,
	logger logging.Logger,
) *MigrationHandlerDeleteGoSSBRepositoryInOldFormat {
	return &MigrationHandlerDeleteGoSSBRepositoryInOldFormat{
		repoReader: repoReader,
		logger:     logger.New("migration_handler_delete_go_ssb_repository_in_old_format"),
	}
}

func (h MigrationHandlerDeleteGoSSBRepositoryInOldFormat) Handle(ctx context.Context, cmd DeleteGoSSBRepositoryInOldFormat) error {
	if cmd.IsZero() {
		return errors.New("zero value of command")
	}

	canRead, err := h.canReadData(ctx, cmd.directory)
	if err != nil {
		return errors.Wrap(err, "error checking if data can be read")
	}

	if !canRead {
		if err := os.RemoveAll(cmd.directory); err != nil {
			return errors.Wrap(err, "error removing the directory")
		}
	}

	return nil
}

func (h MigrationHandlerDeleteGoSSBRepositoryInOldFormat) canReadData(ctx context.Context, dir string) (bool, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	messages, err := h.repoReader.GetMessages(ctx, dir, nil)
	if err != nil {
		return false, errors.Wrap(err, "error getting messages")
	}

	msgOrError, ok := <-messages
	if !ok {
		return false, nil
	}

	if err := msgOrError.Err; err != nil {
		h.logger.WithError(err).Debug("received an error")
		return false, nil
	}

	return true, nil
}
