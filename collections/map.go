package collections

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
