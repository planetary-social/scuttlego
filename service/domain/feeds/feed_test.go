package feeds_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestFeed_AppendMessage_FirstMessageMustBeARootMessage(t *testing.T) {
	feed := fixtures.SomeRefFeed()
	author := fixtures.SomeRefIdentity()

	msg := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		internal.Ptr(fixtures.SomeRefMessage()),
		message.MustNewSequence(2),
		author,
		feed,
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		fixtures.SomeRawMessage(),
	)

	f := feeds.NewFeed(nil)

	err := f.AppendMessage(msg)
	require.EqualError(t, err, "first message in the feed must be a root message")

	msgsToPersist := f.PopForPersisting()
	require.Empty(t, msgsToPersist)
}

func TestFeed_AppendMessage_SubsequentMessagesAreValidated(t *testing.T) {
	firstMessage := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		nil,
		message.MustNewSequence(1),
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefFeed(),
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		fixtures.SomeRawMessage(),
	)

	testCases := []struct {
		Name          string
		Message       message.Message
		ExpectedError func(firstMessage, secondMessage message.Message) error
		ShouldBeAdded bool
	}{
		{
			Name: "valid",
			Message: message.MustNewMessage(
				fixtures.SomeRefMessage(),
				internal.Ptr(firstMessage.Id()),
				message.MustNewSequence(2),
				firstMessage.Author(),
				firstMessage.Feed(),
				fixtures.SomeTime(),
				fixtures.SomeContent(),
				fixtures.SomeRawMessage(),
			),
			ExpectedError: nil,
			ShouldBeAdded: true,
		},
		{
			Name:          "idempotency",
			Message:       firstMessage,
			ExpectedError: nil,
			ShouldBeAdded: false,
		},
		{
			Name: "invalid_author",
			Message: message.MustNewMessage(
				fixtures.SomeRefMessage(),
				internal.Ptr(firstMessage.Id()),
				message.MustNewSequence(2),
				fixtures.SomeRefIdentity(),
				firstMessage.Feed(),
				fixtures.SomeTime(),
				fixtures.SomeContent(),
				fixtures.SomeRawMessage(),
			),
			ExpectedError: func(firstMessage, secondMessage message.Message) error {
				return errors.New("invalid author")
			},
			ShouldBeAdded: false,
		},
		{
			Name: "invalid_feed",
			Message: message.MustNewMessage(
				fixtures.SomeRefMessage(),
				internal.Ptr(firstMessage.Id()),
				message.MustNewSequence(2),
				firstMessage.Author(),
				fixtures.SomeRefFeed(),
				fixtures.SomeTime(),
				fixtures.SomeContent(),
				fixtures.SomeRawMessage(),
			),
			ExpectedError: func(firstMessage, secondMessage message.Message) error {
				return errors.New("invalid feed")
			},
			ShouldBeAdded: false,
		},
		{
			Name: "message_sequence_is_not_right_after_previous_message",
			Message: message.MustNewMessage(
				fixtures.SomeRefMessage(),
				internal.Ptr(firstMessage.Id()),
				message.MustNewSequence(3),
				firstMessage.Author(),
				firstMessage.Feed(),
				fixtures.SomeTime(),
				fixtures.SomeContent(),
				fixtures.SomeRawMessage(),
			),
			ExpectedError: func(firstMessage, secondMessage message.Message) error {
				return fmt.Errorf("this is not the next message in this feed (%s -> %s)", firstMessage, secondMessage)
			},
			ShouldBeAdded: false,
		},
		{
			Name: "previous_message_is_different",
			Message: message.MustNewMessage(
				fixtures.SomeRefMessage(),
				internal.Ptr(fixtures.SomeRefMessage()),
				message.MustNewSequence(2),
				firstMessage.Author(),
				firstMessage.Feed(),
				fixtures.SomeTime(),
				fixtures.SomeContent(),
				fixtures.SomeRawMessage(),
			),
			ExpectedError: func(firstMessage, secondMessage message.Message) error {
				return fmt.Errorf("this is not the next message in this feed (%s -> %s)", firstMessage, secondMessage)
			},
			ShouldBeAdded: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			f := feeds.NewFeed(nil)

			err := f.AppendMessage(firstMessage)
			require.NoError(t, err)

			err = f.AppendMessage(testCase.Message)
			if testCase.ExpectedError != nil {
				expectedErr := testCase.ExpectedError(firstMessage, testCase.Message)
				require.EqualError(t, err, expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			msgsToPersist := f.PopForPersisting()
			if testCase.ShouldBeAdded {
				require.Equal(
					t,
					msgsToPersist,
					[]feeds.MessageToPersist{
						feeds.MustNewMessageToPersist(firstMessage, nil, nil, nil),
						feeds.MustNewMessageToPersist(testCase.Message, nil, nil, nil),
					},
				)
			} else {
				require.Equal(
					t,
					msgsToPersist,
					[]feeds.MessageToPersist{
						feeds.MustNewMessageToPersist(firstMessage, nil, nil, nil),
					},
				)
			}
		})
	}
}

func TestFeed_MessagesWithKnownContentAreCorrectlyRecognized(t *testing.T) {
	msgId := fixtures.SomeRefMessage()
	authorId := fixtures.SomeRefIdentity()
	feedId := fixtures.SomeRefFeed()

	someIdentity := fixtures.SomeRefIdentity()
	someBlob := fixtures.SomeRefBlob()

	testCases := []struct {
		Name             string
		Content          message.Content
		ExpectedContacts []feeds.ContactToSave
		ExpectedPubs     []feeds.PubToSave
		ExpectedBlobs    []feeds.BlobToSave
	}{
		{
			Name: "known_contact",
			Content: message.MustNewContent(
				fixtures.SomeRawContent(),
				known.MustNewContact(
					someIdentity,
					known.MustNewContactActions([]known.ContactAction{known.ContactActionFollow}),
				),
				nil,
			),
			ExpectedContacts: []feeds.ContactToSave{
				feeds.NewContactToSave(
					authorId,
					known.MustNewContact(
						someIdentity,
						known.MustNewContactActions([]known.ContactAction{known.ContactActionFollow}),
					),
				),
			},
		},
		{
			Name: "known_pub",
			Content: message.MustNewContent(
				fixtures.SomeRawContent(),
				known.MustNewPub(
					someIdentity,
					"host",
					1234,
				),
				nil,
			),
			ExpectedPubs: []feeds.PubToSave{
				feeds.NewPubToSave(
					authorId,
					msgId,
					known.MustNewPub(
						someIdentity,
						"host",
						1234,
					),
				),
			},
		},
		{
			Name: "blobs",
			Content: message.MustNewContent(
				fixtures.SomeRawContent(),
				nil,
				[]refs.Blob{
					someBlob,
				},
			),
			ExpectedBlobs: []feeds.BlobToSave{
				feeds.MustNewBlobToSave(someBlob),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			msg := message.MustNewMessage(
				msgId,
				nil,
				message.MustNewSequence(1),
				authorId,
				feedId,
				fixtures.SomeTime(),
				testCase.Content,
				fixtures.SomeRawMessage(),
			)

			t.Run("append", func(t *testing.T) {
				f := feeds.NewFeed(nil)

				err := f.AppendMessage(msg)
				require.NoError(t, err)

				msgsToPersist := f.PopForPersisting()
				require.Equal(
					t,
					msgsToPersist,
					[]feeds.MessageToPersist{
						feeds.MustNewMessageToPersist(
							msg,
							testCase.ExpectedContacts,
							testCase.ExpectedPubs,
							testCase.ExpectedBlobs),
					},
				)
			})

			t.Run("create", func(t *testing.T) {
				format := newFormatMock()
				format.SignResult = msg

				f := feeds.NewFeed(format)

				_, err := f.CreateMessage(fixtures.SomeRawContent(), fixtures.SomeTime(), fixtures.SomePrivateIdentity())
				require.NoError(t, err)

				msgsToPersist := f.PopForPersisting()
				require.Equal(
					t,
					msgsToPersist,
					[]feeds.MessageToPersist{
						feeds.MustNewMessageToPersist(
							msg,
							testCase.ExpectedContacts,
							testCase.ExpectedPubs,
							testCase.ExpectedBlobs),
					},
				)
			})
		})
	}
}

func TestFeed_CreateMessage(t *testing.T) {
	testCases := []struct {
		Name string

		Content   message.RawContent
		Timestamp time.Time
		Private   identity.Private

		ExpectedError error
	}{
		{
			Name:          "valid",
			Content:       fixtures.SomeRawContent(),
			Timestamp:     fixtures.SomeTime(),
			Private:       fixtures.SomePrivateIdentity(),
			ExpectedError: nil,
		},
		{
			Name:          "zero_value_of_raw_message_content",
			Content:       message.RawContent{},
			Timestamp:     fixtures.SomeTime(),
			Private:       fixtures.SomePrivateIdentity(),
			ExpectedError: errors.New("zero value of raw message content"),
		},
		{
			Name:          "zero_value_of_time",
			Content:       fixtures.SomeRawContent(),
			Timestamp:     time.Time{},
			Private:       fixtures.SomePrivateIdentity(),
			ExpectedError: errors.New("zero value of timestamp"),
		},
		{
			Name:          "zero_value_of_private_identity",
			Content:       fixtures.SomeRawContent(),
			Timestamp:     fixtures.SomeTime(),
			Private:       identity.Private{},
			ExpectedError: errors.New("zero value of private identity"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			format := newFormatMock()
			f := feeds.NewFeed(format)

			_, err := f.CreateMessage(testCase.Content, testCase.Timestamp, testCase.Private)
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFeed_CreateMessage_PassingIdentityWhichDoesNotMatchPreviousIdentityIsInvalid(t *testing.T) {
	format := newFormatMock()
	f := feeds.NewFeed(format)

	firstMessage := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		nil,
		message.MustNewSequence(1),
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefFeed(),
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		fixtures.SomeRawMessage(),
	)

	err := f.AppendMessage(firstMessage)
	require.NoError(t, err)

	_, err = f.CreateMessage(fixtures.SomeRawContent(), fixtures.SomeTime(), fixtures.SomePrivateIdentity())
	require.EqualError(t, err, "private identity doesn't match this feed's public identity")
}

func TestFeed_PopForPersistingClearsTheListOfMessagesToPersist(t *testing.T) {
	format := newFormatMock()
	f := feeds.NewFeed(format)

	firstMessage := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		nil,
		message.MustNewSequence(1),
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefFeed(),
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		fixtures.SomeRawMessage(),
	)

	err := f.AppendMessage(firstMessage)
	require.NoError(t, err)

	require.NotEmpty(t, f.PopForPersisting())
	require.Empty(t, f.PopForPersisting())
}

type formatMock struct {
	SignResult message.Message
}

func (f formatMock) Peek(raw message.RawMessage) (feeds.PeekedMessage, error) {
	//TODO implement me
	panic("implement me")
}

func newFormatMock() *formatMock {
	return &formatMock{}
}

func (f formatMock) Verify(raw message.RawMessage) (message.Message, error) {
	return message.Message{}, errors.New("not implemented")
}

func (f formatMock) Load(raw message.VerifiedRawMessage) (message.MessageWithoutId, error) {
	return message.MessageWithoutId{}, errors.New("not implemented")
}

func (f formatMock) Sign(unsigned message.UnsignedMessage, private identity.Private) (message.Message, error) {
	if f.SignResult.IsZero() {
		return fixtures.SomeMessage(unsigned.Sequence(), unsigned.Feed()), nil
	} else {
		return f.SignResult, nil
	}
}
