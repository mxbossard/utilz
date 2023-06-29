package collections

import "log"

func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice))

	for _, element := range slice {
		if predicate(element) {
			result = append(result, element)
		}
	}

	return result
}

func Map[T, I any](items []I, mapper func(I) T) (result []T) {
	if mapper == nil {
		log.Fatal("No mapper func supplied !")
	}
	for _, item := range items {
		result = append(result, mapper(item))
	}
	return
}

func Reduce[T any](items []T, reducer func(T, T) T) (result T) {
	if reducer == nil {
		log.Fatal("No reducer func supplied !")
	}

	result = items[0]
	for _, item := range items[1:] {
		result = reducer(result, item)
	}
	return
}
