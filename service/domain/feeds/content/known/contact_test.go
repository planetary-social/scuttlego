package known

import (
	"fmt"
	"testing"

	"github.com/boreq/errors"
	"github.com/stretchr/testify/require"
)

func TestContactActions(t *testing.T) {
	testCases := []struct {
		Actions       []ContactAction
		ExpectedError error
	}{
		{
			Actions:       []ContactAction{},
			ExpectedError: errors.New("actions can not be empty"),
		},
		{
			Actions:       nil,
			ExpectedError: errors.New("actions can not be empty"),
		},
		{
			Actions: []ContactAction{
				{},
			},
			ExpectedError: errors.New("zero value of action"),
		},
		{
			Actions: []ContactAction{
				ContactActionFollow,
				ContactActionFollow,
			},
			ExpectedError: errors.New("duplicate action: '{follow}'"),
		},
		{
			Actions: []ContactAction{
				ContactActionFollow,
			},
			ExpectedError: nil,
		},
		{
			Actions: []ContactAction{
				ContactActionUnfollow,
			},
			ExpectedError: nil,
		},
		{
			Actions: []ContactAction{
				ContactActionBlock,
			},
			ExpectedError: nil,
		},
		{
			Actions: []ContactAction{
				ContactActionUnblock,
			},
			ExpectedError: nil,
		},
		{
			Actions: []ContactAction{
				ContactActionFollow,
				ContactActionUnfollow,
			},
			ExpectedError: errors.New("both follow and unfollow are present"),
		},
		{
			Actions: []ContactAction{
				ContactActionBlock,
				ContactActionUnblock,
			},
			ExpectedError: errors.New("both block and unblock are present"),
		},
		{
			Actions: []ContactAction{
				ContactActionFollow,
				ContactActionBlock,
			},
			ExpectedError: errors.New("both follow and block are present"),
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%#v", testCase.Actions), func(t *testing.T) {
			actions, err := NewContactActions(testCase.Actions)
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, len(testCase.Actions), len(actions.List()))
			}
		})
	}
}
