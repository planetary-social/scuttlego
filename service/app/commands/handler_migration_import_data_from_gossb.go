package commands

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
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
	repoReader GoSSBRepoReader
}

func NewMigrationHandlerImportDataFromGoSSB(repoReader GoSSBRepoReader) *MigrationHandlerImportDataFromGoSSB {
	return &MigrationHandlerImportDataFromGoSSB{repoReader: repoReader}
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
	}

	return nil
}
