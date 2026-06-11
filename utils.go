package main

func filter[T any](arr []T, predicate func(T) bool) []T {
	out := make([]T, 0, len(arr))

	for _, item := range arr {
		if predicate(item) {
			out = append(out, item)
		}
	}

	return out
}
