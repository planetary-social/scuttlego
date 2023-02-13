package message_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/stretchr/testify/require"
)

func TestSequence(t *testing.T) {
	testCases := []struct {
		Name          string
		Seq           int
		ExpectedError error
	}{
		{
			Name:          "negative_values_are_invalid",
			Seq:           -1,
			ExpectedError: errors.New("sequence must be positive"),
		},
		{
			Name:          "zero_is_invalid",
			Seq:           0,
			ExpectedError: errors.New("sequence must be positive"),
		},
		{
			Name:          "positive_values_are_valid",
			Seq:           1,
			ExpectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			seq, err := message.NewSequence(testCase.Seq)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				require.Equal(t, testCase.Seq, seq.Int())
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}

func TestSequence_IsFirst(t *testing.T) {
	seq1, err := message.NewSequence(1)
	require.NoError(t, err)
	require.True(t, seq1.IsFirst())

	seq2, err := message.NewSequence(2)
	require.NoError(t, err)
	require.False(t, seq2.IsFirst())

	first := message.NewFirstSequence()
	require.True(t, first.IsFirst())
}

func TestSequence_ComesDirectlyBefore(t *testing.T) {
	seq1, err := message.NewSequence(1)
	require.NoError(t, err)

	seq2, err := message.NewSequence(2)
	require.NoError(t, err)

	seq3, err := message.NewSequence(3)
	require.NoError(t, err)

	require.True(t, seq1.ComesDirectlyBefore(seq2))
	require.False(t, seq2.ComesDirectlyBefore(seq1))

	require.True(t, seq2.ComesDirectlyBefore(seq3))
	require.False(t, seq3.ComesDirectlyBefore(seq2))

	require.False(t, seq1.ComesDirectlyBefore(seq3))
	require.False(t, seq3.ComesDirectlyBefore(seq1))
}

func TestSequence_ComesAfter(t *testing.T) {
	seq1, err := message.NewSequence(1)
	require.NoError(t, err)

	seq2, err := message.NewSequence(2)
	require.NoError(t, err)

	seq3, err := message.NewSequence(3)
	require.NoError(t, err)

	require.False(t, seq1.ComesAfter(seq2))
	require.True(t, seq2.ComesAfter(seq1))

	require.False(t, seq2.ComesAfter(seq3))
	require.True(t, seq3.ComesAfter(seq2))

	require.False(t, seq1.ComesAfter(seq3))
	require.True(t, seq3.ComesAfter(seq1))
}

func TestSequence_Next(t *testing.T) {
	seq1, err := message.NewSequence(1)
	require.NoError(t, err)

	seq2, err := message.NewSequence(2)
	require.NoError(t, err)

	require.True(t, seq1.Next() == seq2)
}

func TestSequence_IsZero(t *testing.T) {
	seq1, err := message.NewSequence(1)
	require.NoError(t, err)
	require.False(t, seq1.IsZero())

	require.True(t, message.Sequence{}.IsZero())
}

func TestSequence_Previous_ReturnsFalseIfThereIsNoPreviousSequence(t *testing.T) {
	firstSequence := message.NewFirstSequence()
	_, ok := firstSequence.Previous()
	require.False(t, ok)
}

func TestSequence_Previous_ReturnsPreviousSequenceIfThereIsPreviousSequence(t *testing.T) {
	v, ok := message.MustNewSequence(123).Previous()
	require.True(t, ok)
	require.Equal(t, message.MustNewSequence(122), v)
}
