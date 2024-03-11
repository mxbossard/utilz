package utilz

import (
	"os"
	"strings"
)

func EnvValue(key string) (ok bool, value string) {
	for _, env := range os.Environ() {

		if ok = strings.HasPrefix(env, key+"="); ok {
			splitted := strings.Split(env, "=")
			value = strings.Join(splitted[1:], "")
			return
		}
	}
	return
}
