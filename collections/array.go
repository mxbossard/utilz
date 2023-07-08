package collections

import (
	"log"
	"reflect"
)

type Predicater[T any] interface {
	Predicate(...T) bool
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice))

	for _, element := range slice {
		if predicate(element) {
			result = append(result, element)
		}
	}

	return result
}

func Map[I, T any](items []I, mapper func(I) T) (result []T) {
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

func ContainsAny[T any](slice *[]T, item T) bool {
	for _, i := range *slice {
		if reflect.DeepEqual(i, item) {
			return true
		}
	}
	return false
}

func Contains[T comparable](slice *[]T, item T) bool {
	for _, i := range *slice {
		if i == item {
			return true
		}
	}
	return false
}
