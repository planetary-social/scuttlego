package commands_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
	gossbrefs "go.mindeco.de/ssb-refs"
)

func TestMigrationHandlerImportDataFromGoSSB_MessageReturnedFromRepoReaderIsSaved(t *testing.T) {
	tc, err := di.BuildTestCommands(t)
	require.NoError(t, err)

	directory := fixtures.SomeString()
	receiveLogSequence := fixtures.SomeReceiveLogSequence()
	saveResumeAfterSequenceFn := newSaveResumeAfterSequenceFnMock()

	cmd, err := commands.NewImportDataFromGoSSB(directory, &receiveLogSequence, saveResumeAfterSequenceFn.Fn)
	require.NoError(t, err)

	msgReceiveLogSequence := fixtures.SomeReceiveLogSequence()
	msg := mockGoSSBMessage(t)

	tc.GoSSBRepoReader.MockGetMessages(
		[]commands.GoSSBMessageOrError{
			{
				Value: commands.GoSSBMessage{
					ReceiveLogSequence: msgReceiveLogSequence,
					Message:            msg,
				},
				Err: nil,
			},
		},
	)

	ctx := fixtures.TestContext(t)
	err = tc.MigrationImportDataFromGoSSB.Handle(ctx, cmd)
	require.NoError(t, err)

	require.Equal(t,
		[]mocks.GoSSBRepoReaderMockGetMessagesCall{
			{
				Directory:           directory,
				ResumeAfterSequence: &receiveLogSequence,
			},
		},
		tc.GoSSBRepoReader.GoSSBRepoReaderMockGetMessagesCalls,
	)

	require.Equal(
		t,
		[]mocks.FeedRepositoryMockUpdateFeedIgnoringReceiveLogCall{
			{
				Feed: refs.MustNewFeed(msg.Author().Sigil()),
				MessagesToPersist: []refs.Message{
					refs.MustNewMessage(msg.Key().String()),
				},
			},
		},
		tc.FeedRepository.UpdateFeedIgnoringReceiveLogCalls(),
	)

	require.Equal(
		t,
		[]mocks.ReceiveLogRepositoryPutUnderSpecificSequenceCall{
			{
				Id:       refs.MustNewMessage(msg.Key().String()),
				Sequence: msgReceiveLogSequence,
			},
		},
		tc.ReceiveLog.PutUnderSpecificSequenceCalls,
	)

	require.Equal(t,
		[]common.ReceiveLogSequence{
			msgReceiveLogSequence,
		},
		saveResumeAfterSequenceFn.calls,
	)
}

func mockGoSSBMessage(t *testing.T) gossbrefs.Message {
	key, err := gossbrefs.ParseMessageRef(fixtures.SomeRefMessage().String())
	require.NoError(t, err)

	author, err := gossbrefs.ParseFeedRef(fixtures.SomeRefIdentity().String())
	require.NoError(t, err)

	return mockMessage{
		key:    key,
		author: author,
	}
}

type saveResumeAfterSequenceFnMock struct {
	calls []common.ReceiveLogSequence
}

func newSaveResumeAfterSequenceFnMock() *saveResumeAfterSequenceFnMock {
	return &saveResumeAfterSequenceFnMock{}
}

func (m *saveResumeAfterSequenceFnMock) Fn(s common.ReceiveLogSequence) error {
	m.calls = append(m.calls, s)
	return nil
}

type mockMessage struct {
	key    gossbrefs.MessageRef
	author gossbrefs.FeedRef
}

func (m mockMessage) Key() gossbrefs.MessageRef {
	return m.key
}

func (m mockMessage) Previous() *gossbrefs.MessageRef {
	return nil
}

func (m mockMessage) Seq() int64 {
	return 1
}

func (m mockMessage) Claimed() time.Time {
	return fixtures.SomeTime()
}

func (m mockMessage) Received() time.Time {
	return fixtures.SomeTime()
}

func (m mockMessage) Author() gossbrefs.FeedRef {
	return m.author
}

func (m mockMessage) ContentBytes() []byte {
	return fixtures.SomeBytes()
}

func (m mockMessage) ValueContent() *gossbrefs.Value {
	return nil
}

func (m mockMessage) ValueContentJSON() json.RawMessage {
	return fixtures.SomeRawMessage().Bytes()
}
