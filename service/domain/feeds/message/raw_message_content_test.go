package message_test

import (
	"errors"
	"testing"

	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
)

func TestRawMessageContent_NewRawMessageContent(t *testing.T) {
	testCases := []struct {
		Name          string
		Slice         []byte
		ExpectedError error
	}{
		{
			Name:          "empty_slice_is_invalid",
			Slice:         nil,
			ExpectedError: errors.New("empty content"),
		},
		{
			Name:          "empty_slice_is_invalid",
			Slice:         []byte{},
			ExpectedError: errors.New("empty content"),
		},
		{
			Name:          "non_empty_slice_is_valid",
			Slice:         []byte{1, 2, 3},
			ExpectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := message.NewRawMessageContent(testCase.Slice)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}

func TestRawMessageContent_NewRawMessageContent_CopiesSlice(t *testing.T) {
	input := []byte{1, 2, 3}
	rawMessage, err := message.NewRawMessageContent(input)
	require.NoError(t, err)
	input[0] = 42
	require.NotEqual(t, input, rawMessage.Bytes())
}

func TestRawMessageContent_Bytes_CopiesSlice(t *testing.T) {
	rawMessage, err := message.NewRawMessageContent([]byte{1, 2, 3})
	require.NoError(t, err)
	output := rawMessage.Bytes()
	output[0] = 42
	require.NotEqual(t, output, rawMessage.Bytes())
}

func TestRawMessageContent_Bytes(t *testing.T) {
	input := []byte{1, 2, 3}
	rawMessage, err := message.NewRawMessageContent(input)
	require.NoError(t, err)
	require.Equal(t, input, rawMessage.Bytes())
}

func TestRawMessageContent_IsZero(t *testing.T) {
	rawMessage, err := message.NewRawMessageContent([]byte{1, 2, 3})
	require.NoError(t, err)
	require.False(t, rawMessage.IsZero())
	require.True(t, message.RawMessageContent{}.IsZero())
}
