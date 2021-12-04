package trust

import (
	//"fmt"
	//"os"
	//"path/filepath"
	"crypto/sha256"
	"encoding/hex"
	//"errors"

	"golang.org/x/mod/sumdb/dirhash"
)

func SignDirAll(path string) (sign string, err error) {
	sign, err = dirhash.HashDir(path, "", dirhash.Hash1)
	if err != nil {
		return
	}
	return
}

func HashPath(path string) (h string) {
	hBytes := sha256.Sum256([]byte(path))
	h = hex.EncodeToString(hBytes[:])
	return
}

