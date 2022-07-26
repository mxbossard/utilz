package collections

func filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice))

	for _, element := range slice {
		if predicate(element) {
			result = append(result, element)
		}
	}

	return result
}
