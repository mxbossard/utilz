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

func MapKeys[K comparable, V any](m map[K]V) (keys []K) {
	for k := range m {
		keys = append(keys, k)
	}
	return
}

func MapOrderedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := MapKeys(m)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func MapValues[K comparable, V any](m map[K]V) (values []V) {
	for _, v := range m {
		values = append(values, v)
	}
	return
}

func MapOrderedValues[K cmp.Ordered, V any](m map[K]V) (values []V) {
	keys := MapOrderedKeys(m)
	for _, k := range keys {
		values = append(values, m[k])
	}
	return
}
