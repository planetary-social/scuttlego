package known_test

import (
	"fmt"
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/stretchr/testify/require"
)

func TestContactActions(t *testing.T) {
	testCases := []struct {
		Actions       []known.ContactAction
		ExpectedError error
	}{
		{
			Actions:       []known.ContactAction{},
			ExpectedError: errors.New("actions can not be empty"),
		},
		{
			Actions:       nil,
			ExpectedError: errors.New("actions can not be empty"),
		},
		{
			Actions: []known.ContactAction{
				{},
			},
			ExpectedError: errors.New("zero value of action"),
		},
		{
			Actions: []known.ContactAction{
				known.ContactActionFollow,
				known.ContactActionFollow,
			},
			ExpectedError: errors.New("duplicate action: '{follow}'"),
		},
		{
			Actions: []known.ContactAction{
				known.ContactActionFollow,
			},
			ExpectedError: nil,
		},
		{
			Actions: []known.ContactAction{
				known.ContactActionUnfollow,
			},
			ExpectedError: nil,
		},
		{
			Actions: []known.ContactAction{
				known.ContactActionBlock,
			},
			ExpectedError: nil,
		},
		{
			Actions: []known.ContactAction{
				known.ContactActionUnblock,
			},
			ExpectedError: nil,
		},
		{
			Actions: []known.ContactAction{
				known.ContactActionFollow,
				known.ContactActionUnfollow,
			},
			ExpectedError: errors.New("both follow and unfollow are present"),
		},
		{
			Actions: []known.ContactAction{
				known.ContactActionBlock,
				known.ContactActionUnblock,
			},
			ExpectedError: errors.New("both block and unblock are present"),
		},
		{
			Actions: []known.ContactAction{
				known.ContactActionFollow,
				known.ContactActionBlock,
			},
			ExpectedError: errors.New("both follow and block are present"),
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%#v", testCase.Actions), func(t *testing.T) {
			actions, err := known.NewContactActions(testCase.Actions)
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, len(testCase.Actions), len(actions.List()))
			}
		})
	}
}
