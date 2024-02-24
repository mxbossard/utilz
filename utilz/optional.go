package utilz

import "log"

type Optional[T any] struct {
	value *T
}

func (o Optional[T]) Get() T {
	if o.IsEmpty() {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Fatal("attempt to get empty optional value")
	}
	return *o.value
}

func (o Optional[T]) IsEmpty() bool {
	return o.value == nil
}

func (o Optional[T]) IsPresent() bool {
	return o.value != nil
}

func (o Optional[T]) IfPresent(f func(T) error) error {
	if o.IsPresent() {
		return f(*o.value)
	}
	return nil
}

func Empty[T any]() Optional[T] {
	return Optional[T]{value: nil}
}

func Of[T any](value T) Optional[T] {
	return Optional[T]{value: &value}
}
