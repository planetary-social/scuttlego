package internal

import (
	"math/rand"
	"sort"
)

func SortSlice[T any](slice []T, less func(a, b T) bool) {
	sort.Slice(slice, func(i, j int) bool {
		return less(slice[i], slice[j])
	})
}

func ShuffleSlice[T any](slice []T) {
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

func RandomElement[T any](slice []T) T {
	return slice[rand.Intn(len(slice))]

}
