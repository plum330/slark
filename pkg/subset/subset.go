package subset

import "math/rand"

func Subset[T any](s []T, subset int) []T {
	rand.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})
	if len(s) <= subset {
		return s
	}
	return s[:subset]
}
