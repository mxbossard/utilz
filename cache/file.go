package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	//"errors"
)

var _ = fmt.Printf

type Cache interface {
	LoadString(key string) (value string, ok bool, err error)
	StoreString(key, value string) (err error)
}

type persistentCache struct {
	path string
}

func NewPersistentCache(path string) (cache Cache, err error) {
	path, err = filepath.Abs(path)
	if err != nil {
		return
	}
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return
	}
	cache = persistentCache{path}
	return
}

func (c persistentCache) bucketFilepath(key string) (dir, path string) {
	hashedKey := hashKey(key)
	level1 := hashedKey[:2]
	level2 := hashedKey[2:4]
	dir = filepath.Join(c.path, level1, level2)
	path = filepath.Join(dir, hashedKey)
	return
}

func (c persistentCache) LoadString(key string) (value string, ok bool, err error) {
	_, bucketPath := c.bucketFilepath(key)
	//fmt.Printf("Loading value from bucket: %s\n", bucketPath)
	content, err := os.ReadFile(bucketPath)
	if os.IsNotExist(err) {
		err = nil
		return
	} else if err != nil {
		return
	}

	ok = true
	value = string(content)
	return
}

func (c persistentCache) StoreString(key, value string) (err error) {
	dir, path := c.bucketFilepath(key)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return
	}
	//fmt.Printf("Storing value: %s in bucket: %s ...\n", value, bucket)
	err = os.WriteFile(path, []byte(value), 0644)
	return
}

func hashKey(key string) (h string) {
	hBytes := sha256.Sum256([]byte(key))
	h = hex.EncodeToString(hBytes[:])
	return
}
