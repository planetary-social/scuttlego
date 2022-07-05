package feeds_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	msgcontents "github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestAppend(t *testing.T) {
	feed := fixtures.SomeRefFeed()
	author := fixtures.SomeRefIdentity()

	msg1 := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		nil,
		message.MustNewSequence(1),
		author,
		feed,
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		fixtures.SomeRawMessage(),
	)

	prevId := msg1.Id()

	msg2 := message.MustNewMessage(
		fixtures.SomeRefMessage(),
		&prevId,
		message.MustNewSequence(2),
		author,
		feed,
		fixtures.SomeTime(),
		fixtures.SomeContent(),
		fixtures.SomeRawMessage(),
	)

	f := feeds.NewFeed(nil)

	err := f.AppendMessage(msg1)
	require.NoError(t, err)

	err = f.AppendMessage(msg2)
	require.NoError(t, err)

	msgs, contacts, pubs, blobs := f.PopForPersisting()
	require.Len(t, msgs, 2)
	require.Len(t, contacts, 0)
	require.Len(t, pubs, 0)
	require.Len(t, blobs, 0)
}

func TestAppendMessageWithKnownContent(t *testing.T) {
	msgId := fixtures.SomeRefMessage()
	authorId := fixtures.SomeRefIdentity()
	feedId := fixtures.SomeRefFeed()

	someIdentity := fixtures.SomeRefIdentity()
	someBlob := fixtures.SomeRefBlob()

	testCases := []struct {
		Name             string
		Content          msgcontents.KnownMessageContent
		ExpectedContacts []feeds.ContactToSave
		ExpectedPubs     []feeds.PubToSave
		ExpectedBlobs    []feeds.BlobsToSave
	}{
		{
			Name: "contact",
			Content: msgcontents.MustNewContact(
				someIdentity,
				msgcontents.ContactActionFollow,
			),
			ExpectedContacts: []feeds.ContactToSave{
				feeds.NewContactToSave(
					authorId,
					msgcontents.MustNewContact(
						someIdentity,
						msgcontents.ContactActionFollow,
					),
				),
			},
		},
		{
			Name: "pub",
			Content: msgcontents.MustNewPub(
				someIdentity,
				"host",
				1234,
			),
			ExpectedPubs: []feeds.PubToSave{
				feeds.NewPubToSave(
					authorId,
					msgId,
					msgcontents.MustNewPub(
						someIdentity,
						"host",
						1234,
					),
				),
			},
		},
		{
			Name: "about",
			Content: msgcontents.MustNewAbout(
				&someBlob,
			),
			ExpectedBlobs: []feeds.BlobsToSave{
				feeds.NewBlobsToSave(
					feedId,
					msgId,
					[]refs.Blob{
						someBlob,
					},
				),
			},
		},
		{
			Name: "post",
			Content: msgcontents.MustNewPost(
				[]refs.Blob{someBlob},
			),
			ExpectedBlobs: []feeds.BlobsToSave{
				feeds.NewBlobsToSave(
					feedId,
					msgId,
					[]refs.Blob{
						someBlob,
					},
				),
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

			f := feeds.NewFeed(nil)

			err := f.AppendMessage(msg)
			require.NoError(t, err)

			msgs, contacts, pubs, blobs := f.PopForPersisting()
			require.Len(t, msgs, 1)
			require.Equal(t, testCase.ExpectedContacts, contacts)
			require.Equal(t, testCase.ExpectedPubs, pubs)
			require.Equal(t, testCase.ExpectedBlobs, blobs)
		})
	}
}
