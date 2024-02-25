package utilz

import (
	"fmt"
	"log"
)

type Optional[T comparable] struct {
	value *T
}

func (o Optional[T]) Get() T {
	if o.IsEmpty() {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Fatal("attempt to get empty optional value")
	}
	return *o.value
}

func (o Optional[T]) GetOrError() (val T, err error) {
	if o.IsEmpty() {
		err = fmt.Errorf("attempt to get empty optional value")
		return
	}
	val = *o.value
	return
}

func (o Optional[T]) IsEmpty() bool {
	return o.value == nil
}

func (o Optional[T]) IsPresent() bool {
	return o.value != nil
}

func (o Optional[T]) Is(expected T) bool {
	if !o.IsPresent() {
		return false
	}
	val := *o.value
	return val == expected
}

func (o Optional[T]) IfPresent(f func(T) error) error {
	if o.IsPresent() {
		return f(*o.value)
	}
	return nil
}

func EmptyOptionnal[T comparable]() Optional[T] {
	return Optional[T]{value: nil}
}

func OptionnalOf[T comparable](value T) Optional[T] {
	return Optional[T]{value: &value}
}

type AnyOptional[T any] struct {
	value *T
}

func (o AnyOptional[T]) Get() T {
	if o.IsEmpty() {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Fatal("attempt to get empty optional value")
	}
	return *o.value
}

func (o AnyOptional[T]) GetOrError() (val T, err error) {
	if o.IsEmpty() {
		err = fmt.Errorf("attempt to get empty optional value")
		return
	}
	val = *o.value
	return
}

func (o AnyOptional[T]) IsEmpty() bool {
	return o.value == nil
}

func (o AnyOptional[T]) IsPresent() bool {
	return o.value != nil
}

func (o AnyOptional[T]) IfPresent(f func(T) error) error {
	if o.IsPresent() {
		return f(*o.value)
	}
	return nil
}

func EmptyAnyOptionnal[T any]() AnyOptional[T] {
	return AnyOptional[T]{value: nil}
}

func AnyOptionnalOf[T any](value T) AnyOptional[T] {
	return AnyOptional[T]{value: &value}
}
