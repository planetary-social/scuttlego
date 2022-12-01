package migrations

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
	"go.cryptoscope.co/luigi"
	"go.cryptoscope.co/margaret"
	"go.cryptoscope.co/ssb/sbot"
	refs "go.mindeco.de/ssb-refs"
)

// getMessagesChannelSize ensures that there is some buffering happening to
// reduce switching between goroutines (at lest I think this should help with
// that - no clue how the runtime handles this).
const getMessagesChannelSize = 1000

type GoSSBRepoReader struct {
	networkKey  boxstream.NetworkKey
	messageHMAC formats.MessageHMAC
	logger      logging.Logger
}

func NewGoSSBRepoReader(
	networkKey boxstream.NetworkKey,
	messageHMAC formats.MessageHMAC,
	logger logging.Logger,
) *GoSSBRepoReader {
	return &GoSSBRepoReader{
		networkKey:  networkKey,
		messageHMAC: messageHMAC,
		logger:      logger.New("go_ssb_repo_reader"),
	}
}

func (m GoSSBRepoReader) GetMessages(ctx context.Context, directory string) (<-chan commands.GoSSBMessageOrError, error) {
	_, err := os.Stat(directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ch := make(chan commands.GoSSBMessageOrError)
			close(ch)
			return ch, nil
		}
		return nil, errors.Wrap(err, "failed to stat directory")
	}

	ctx, cancel := context.WithCancel(ctx)

	bot, err := m.createBot(ctx, directory)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "error making a bot")
	}

	m.logger.
		WithField("max_receive_log_sequence", bot.ReceiveLog.Seq()).
		Debug("created the bot")

	src, err := bot.ReceiveLog.Query(
		margaret.SeqWrap(true),
	)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "error querying receive log")
	}

	ch := make(chan commands.GoSSBMessageOrError, getMessagesChannelSize)

	go func() {
		defer cancel()
		defer close(ch)

		for {
			rxLogSeq, msg, err := m.getNextMessage(ctx, src)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}

				select {
				case <-ctx.Done():
					return
				case ch <- commands.GoSSBMessageOrError{
					Err: errors.Wrap(err, "error getting next message"),
				}:
					return
				}
			}

			select {
			case <-ctx.Done():
				return
			case ch <- commands.GoSSBMessageOrError{
				Value: commands.GoSSBMessage{
					ReceiveLogSequence: rxLogSeq,
					Message:            msg,
				},
			}:
				continue
			}
		}
	}()

	return ch, nil
}

func (m GoSSBRepoReader) getNextMessage(ctx context.Context, src luigi.Source) (int64, refs.Message, error) {
	for {
		v, err := src.Next(ctx)
		if err != nil {
			if luigi.IsEOS(err) {
				return 0, nil, io.EOF
			}
			return 0, nil, errors.Wrap(err, "error getting next value")
		}

		if err, ok := v.(error); ok {
			if margaret.IsErrNulled(err) {
				continue
			}
			return 0, nil, errors.Wrap(err, "margaret returned an error")
		}

		sw, ok := v.(margaret.SeqWrapper)
		if !ok {
			return 0, nil, fmt.Errorf("expected message seq wrapper but got '%T'", v)
		}

		rxLogSeq := sw.Seq()

		msg, ok := sw.Value().(refs.Message)
		if !ok {
			return 0, nil, fmt.Errorf("expected message but got '%T'", sw.Value())
		}

		return rxLogSeq, msg, nil
	}
}

func (m GoSSBRepoReader) createBot(ctx context.Context, directory string) (*sbot.Sbot, error) {
	options := []sbot.Option{
		sbot.WithRepoPath(directory),
		sbot.WithContext(ctx),
		sbot.WithAppKey(m.networkKey.Bytes()),
	}
	if !m.messageHMAC.IsZero() {
		options = append(options, sbot.WithHMACSigning(m.messageHMAC.Bytes()))
	}

	bot, err := sbot.New(options...)
	if err != nil {
		return nil, errors.Wrap(err, "error calling sbot new")
	}

	return bot, nil
}
