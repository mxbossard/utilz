package collections

import (
	"cmp"
	"sort"
)

func CloneMap(m map[string]any) map[string]any {
	cp := make(map[string]any)
	for k, v := range m {
		vm, ok := v.(map[string]any)
		if ok {
			cp[k] = CloneMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}

// FIXME: remove
func mapKeys[K comparable, V any](m map[K]V) (keys []K) {
	for k := range m {
		keys = append(keys, k)
	}
	return
}

// FIXME: remove
func mapOrderedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := mapKeys(m)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

// FIXME: remove
func mapValues[K comparable, V any](m map[K]V) (values []V) {
	for _, v := range m {
		values = append(values, v)
	}
	return
}

// FIXME: remove
func mapOrderedValues[K cmp.Ordered, V any](m map[K]V) (values []V) {
	keys := mapOrderedKeys(m)
	for _, k := range keys {
		values = append(values, m[k])
	}
	return
}

func Keys[K comparable, V any](in map[K]V) []K {
	values := make([]K, 0, len(in))
	for k, _ := range in {
		values = append(values, k)
	}
	return values
}

func OrderedKeys[K cmp.Ordered, V any](in map[K]V) []K {
	keys := Keys(in)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func Values[K comparable, V any](in map[K]V) []V {
	values := make([]V, 0, len(in))
	for _, v := range in {
		values = append(values, v)
	}
	return values
}

func OrderedValues[K cmp.Ordered, V any](in map[K]V) []V {
	values := make([]V, 0, len(in))
	keys := OrderedKeys(in)
	for _, k := range keys {
		v := in[k]
		values = append(values, v)
	}
	return values
}
