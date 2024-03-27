package utilz

import (
	"fmt"
)

func EmptyAnyOptional[T any]() AnyOptional[T] {
	return AnyOptional[T]{Value: nil}
}

func AnyOptionalOf[T any](value T) AnyOptional[T] {
	return AnyOptional[T]{Value: &value}
}

func EmptyOptional[T comparable]() Optional[T] {
	return Optional[T]{AnyOptional: EmptyAnyOptional[T]()}
}

func OptionalOf[T comparable](value T) Optional[T] {
	return Optional[T]{AnyOptional: AnyOptionalOf(value)}
}

type AnyOptional[T any] struct {
	Value *T //`yaml:""`
	Def   *T //`yaml:""`
}

func (o AnyOptional[T]) MarshalYAML() (any, error) {
	anonymous := struct {
		Value *T
		Def   *T
	}{
		Value: o.Value,
		Def:   o.Def,
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
	o.Value = anonymous.Value
	o.Def = anonymous.Def
	return nil
}

func (o AnyOptional[T]) String() string {
	if o.IsPresent() {
		return fmt.Sprint(o.Get())
	}
	return "<EMPTY>"
}

func (o *AnyOptional[T]) Clear() {
	o.Value = nil
	o.Def = nil
}

func (o AnyOptional[T]) GetOrError() (val T, err error) {
	if o.IsEmpty() {
		err = fmt.Errorf("attempt to get empty optional value")
		return
	}
	if o.Value != nil {
		val = *o.Value
	} else if o.Def != nil {
		val = *o.Def
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
	o.Value = &val
}

func (o *AnyOptional[T]) Default(val T) {
	o.Def = &val
}

func (o *AnyOptional[T]) Merge(right AnyOptional[T]) {
	if right.IsPresent() {
		o.Set(right.Get())
	}
}

func (o AnyOptional[T]) IsEmpty() bool {
	return o.Value == nil && o.Def == nil
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
