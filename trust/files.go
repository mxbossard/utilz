package trust

import (
	//"fmt"
	//"os"
	//"path/filepath"
	//"errors"

	"golang.org/x/mod/sumdb/dirhash"
)

func SignDirContent(path string) (sign string, err error) {
	sign, err = dirhash.HashDir(path, "", dirhash.Hash1)
	if err != nil {
		return
	}
	return
}

