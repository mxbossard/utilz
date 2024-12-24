package collections

import (
	"log"
	"reflect"
)

type Predicater[T any] interface {
	Predicate(...T) bool
}

func Filter[T any](items *[]T, predicate func(T) bool) []T {
	if items == nil {
		log.Fatal("No collection supplied !")
	}
	result := make([]T, 0, len(*items))

	for _, element := range *items {
		if predicate(element) {
			result = append(result, element)
		}
	}

	return result
}

func Map[I, T any](items *[]I, mapper func(I) T) (result []T) {
	if items == nil {
		log.Fatal("No collection supplied !")
	}
	if mapper == nil {
		log.Fatal("No mapper func supplied !")
	}
	for _, item := range *items {
		result = append(result, mapper(item))
	}
	return
}

func Reduce[T any](items *[]T, reducer func(T, T) T) (result T) {
	if items == nil {
		log.Fatal("No collection supplied !")
	}
	if reducer == nil {
		log.Fatal("No reducer func supplied !")
	}

	result = (*items)[0]
	for _, item := range (*items)[1:] {
		result = reducer(result, item)
	}
	return
}

func ContainsAny[T any](slice *[]T, item T) bool {
	if slice == nil {
		log.Fatal("No collection supplied !")
	}
	for _, i := range *slice {
		if reflect.DeepEqual(i, item) {
			return true
		}
	}
	return false
}

func Contains[T comparable](slice *[]T, item T) bool {
	if slice == nil {
		log.Fatal("No collection supplied !")
	}
	for _, i := range *slice {
		if i == item {
			return true
		}
	}
	return false
}

func Match[T comparable](slice1, slice2 *[]T) bool {
	// TODO: match if 2 slices contains same elements unordered
	return false
}

func ExactMatch[T comparable](slice1, slice2 *[]T) bool {
	// TODO: match if 2 slices contains same elements in same order
	return false
}

func CloneSliceReflect(s any) any {
	t, v := reflect.TypeOf(s), reflect.ValueOf(s)
	c := reflect.MakeSlice(t, v.Len(), v.Len())
	reflect.Copy(c, v)
	return c.Interface()
}

// return left items not in right
func KeepLeft[T any](left, right *[]T) (res []T) {
	for _, l := range *left {
		if !ContainsAny[T](right, l) {
			res = append(res, l)
		}
	}
	return
}

// return items in left and right
func Intersect[T any](left, right *[]T) (res []T) {
	for _, l := range *left {
		if ContainsAny[T](right, l) {
			res = append(res, l)
		}
	}
	return
}

func Deduplicate[T comparable](slices ...*[]T) []T {
	mapSet := make(map[T]bool)
	for _, slice := range slices {
		for _, item := range *slice {
			mapSet[item] = true
		}
	}
	return Keys(mapSet)
}

