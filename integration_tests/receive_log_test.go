package integration_tests

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestReceiveLogDoesNotBreakWhenFeedsAreBannedAndThereforeRemoved(t *testing.T) {
	ts, err := di.BuildIntegrationTestsService(t)
	require.NoError(t, err)

	// publish messages
	iden1 := fixtures.SomePrivateIdentity()
	msgRef1, err := publishAs(ts, iden1)
	require.NoError(t, err)

	iden2 := fixtures.SomePrivateIdentity()
	msgRef2, err := publishAs(ts, iden2)
	require.NoError(t, err)

	iden3 := fixtures.SomePrivateIdentity()
	msgRef3, err := publishAs(ts, iden3)
	require.NoError(t, err)

	getMessageQuery, err := queries.NewGetMessage(msgRef1)
	require.NoError(t, err)

	// get messages
	msg1, err := ts.Service.App.Queries.GetMessage.Handle(getMessageQuery)
	require.NoError(t, err)

	getMessageQuery, err = queries.NewGetMessage(msgRef2)
	require.NoError(t, err)

	msg2, err := ts.Service.App.Queries.GetMessage.Handle(getMessageQuery)
	require.NoError(t, err)

	getMessageQuery, err = queries.NewGetMessage(msgRef3)
	require.NoError(t, err)

	msg3, err := ts.Service.App.Queries.GetMessage.Handle(getMessageQuery)
	require.NoError(t, err)

	// check receive log
	query, err := queries.NewReceiveLog(common.MustNewReceiveLogSequence(0), 10)
	require.NoError(t, err)

	msgs, err := ts.Service.App.Queries.ReceiveLog.Handle(query)
	require.NoError(t, err)
	require.Equal(t,
		[]queries.LogMessage{
			{
				Message:  msg1,
				Sequence: common.MustNewReceiveLogSequence(0),
			},
			{
				Message:  msg2,
				Sequence: common.MustNewReceiveLogSequence(1),
			},
			{
				Message:  msg3,
				Sequence: common.MustNewReceiveLogSequence(2),
			},
		},
		msgs,
	)

	// ban one feed
	hashFeed2, err := ts.BanListHasher.HashForFeed(msg2.Feed())
	require.NoError(t, err)

	addToBanListCommand, err := commands.NewAddToBanList(hashFeed2)
	require.NoError(t, err)

	err = ts.Service.App.Commands.AddToBanList.Handle(addToBanListCommand)
	require.NoError(t, err)

	// check receive log
	query, err = queries.NewReceiveLog(common.MustNewReceiveLogSequence(0), 10)
	require.NoError(t, err)

	msgs, err = ts.Service.App.Queries.ReceiveLog.Handle(query)
	require.NoError(t, err)
	require.Equal(t,
		[]queries.LogMessage{
			{
				Message:  msg1,
				Sequence: common.MustNewReceiveLogSequence(0),
			},
			{
				Message:  msg3,
				Sequence: common.MustNewReceiveLogSequence(2),
			},
		},
		msgs,
	)
}

func publishAs(ts di.IntegrationTestsService, iden identity.Private) (refs.Message, error) {
	cmd, err := commands.NewPublishRawAsIdentity(fixtures.SomeRawContent().Bytes(), iden)
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "error creating a command")
	}

	ref, err := ts.Service.App.Commands.PublishRawAsIdentity.Handle(cmd)
	if err != nil {
		return refs.Message{}, errors.Wrap(err, "error calling the command")
	}

	return ref, nil
}
