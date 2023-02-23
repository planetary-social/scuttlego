package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueue_DequeueReturnsFalseIfQueueIsEmpty(t *testing.T) {
	q := NewQueue[int]()
	_, ok := q.Dequeue()
	require.False(t, ok)
}

func TestQueue_DequeueReturnsItemsAddedByEnqueue(t *testing.T) {
	q := NewQueue[int]()

	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	v, ok := q.Dequeue()
	require.True(t, ok)
	require.Equal(t, 1, v)

	v, ok = q.Dequeue()
	require.True(t, ok)
	require.Equal(t, 2, v)

	v, ok = q.Dequeue()
	require.True(t, ok)
	require.Equal(t, 3, v)

	_, ok = q.Dequeue()
	require.False(t, ok)
}
