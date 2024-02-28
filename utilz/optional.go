package utilz

import (
	"fmt"
)

func EmptyAnyOptional[T any]() AnyOptional[T] {
	return AnyOptional[T]{value: nil}
}

func AnyOptionalOf[T any](value T) AnyOptional[T] {
	return AnyOptional[T]{value: &value}
}

func EmptyOptional[T comparable]() Optional[T] {
	return Optional[T]{AnyOptional: EmptyAnyOptional[T]()}
}

func OptionalOf[T comparable](value T) Optional[T] {
	return Optional[T]{AnyOptional: AnyOptionalOf(value)}
}

type AnyOptional[T any] struct {
	value *T //`yaml:""`
	def   *T //`yaml:""`
}

func (o AnyOptional[T]) MarshalYAML() (any, error) {
	anonymous := struct {
		Value *T
		Def   *T
	}{
		Value: o.value,
		Def:   o.def,
	}
	return anonymous, nil
}

func (o *AnyOptional[T]) UnmarshalYAML(unmarshal func(any) error) error {
	anonymous := struct {
		Value *T
		Def   *T
	}{}
	err := unmarshal(&anonymous)
	if err != nil {
		return err
	}
	o.value = anonymous.Value
	o.def = anonymous.Def
	return nil
}

func (o AnyOptional[T]) String() string {
	if o.IsPresent() {
		return fmt.Sprint(o.Get())
	}
	return "<EMPTY>"
}

func (o *AnyOptional[T]) Clear() {
	o.value = nil
	o.def = nil
}

func (o AnyOptional[T]) GetOrError() (val T, err error) {
	if o.IsEmpty() {
		err = fmt.Errorf("attempt to get empty optional value")
		return
	}
	if o.value != nil {
		val = *o.value
	} else if o.def != nil {
		val = *o.def
	}
	return
}

func (o AnyOptional[T]) Get() T {
	val, err := o.GetOrError()
	if err != nil {
		panic(err)
	}
	return val
}

func (o AnyOptional[T]) GetOr(def T) T {
	if !o.IsPresent() {
		return def
	}
	return o.Get()
}

func (o *AnyOptional[T]) Set(val T) {
	o.value = &val
}

func (o *AnyOptional[T]) Default(val T) {
	o.def = &val
}

func (o *AnyOptional[T]) Merge(right AnyOptional[T]) {
	if right.IsPresent() {
		o.Set(right.Get())
	}
}

func (o AnyOptional[T]) IsEmpty() bool {
	return o.value == nil && o.def == nil
}

func (o AnyOptional[T]) IsPresent() bool {
	return !o.IsEmpty()
}

func (o AnyOptional[T]) IfPresent(f func(T) error) error {
	if o.IsPresent() {
		return f(o.Get())
	}
	return nil
}

type Optional[T comparable] struct {
	AnyOptional[T] `yaml:",inline"`
}

func (o *Optional[T]) Merge(right Optional[T]) {
	if right.IsPresent() {
		o.Set(right.Get())
	}
}

func (o Optional[T]) Is(expected T) bool {
	if !o.IsPresent() {
		return false
	}
	return o.Get() == expected
}
