package mocks

import (
	"context"

	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/common"
)

type GoSSBRepoReaderMockGetMessagesCall struct {
	Directory           string
	ResumeAfterSequence *common.ReceiveLogSequence
}

type GoSSBRepoReaderMock struct {
	getMessagesReturnValue              []commands.GoSSBMessageOrError
	GoSSBRepoReaderMockGetMessagesCalls []GoSSBRepoReaderMockGetMessagesCall
}

func NewGoSSBRepoReaderMock() *GoSSBRepoReaderMock {
	return &GoSSBRepoReaderMock{}
}

func (g *GoSSBRepoReaderMock) GetMessages(ctx context.Context, directory string, resumeAfterSequence *common.ReceiveLogSequence) (<-chan commands.GoSSBMessageOrError, error) {
	g.GoSSBRepoReaderMockGetMessagesCalls = append(
		g.GoSSBRepoReaderMockGetMessagesCalls,
		GoSSBRepoReaderMockGetMessagesCall{
			Directory:           directory,
			ResumeAfterSequence: resumeAfterSequence,
		},
	)
	ch := make(chan commands.GoSSBMessageOrError)
	go func() {
		defer close(ch)
		for i := range g.getMessagesReturnValue {
			select {
			case ch <- g.getMessagesReturnValue[i]:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func (g *GoSSBRepoReaderMock) MockGetMessages(msgs []commands.GoSSBMessageOrError) {
	g.getMessagesReturnValue = msgs
}
