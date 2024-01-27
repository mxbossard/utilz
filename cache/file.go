package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	//"errors"
)

var _ = fmt.Printf

type Cache[T any] interface {
	Load(key string) (value T, ok bool, err error)
	Store(key string, value T) (err error)
}

type StringCache interface {
	LoadString(key string) (value string, ok bool, err error)
	StoreString(key, value string) (err error)
}

type persistentCache[T any] struct {
	mutex *sync.Mutex
	path  string
}

func NewPersistentCache[T any](path string) (cache Cache[T], err error) {
	path, err = filepath.Abs(path)
	if err != nil {
		return
	}
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return
	}
	var mutex sync.Mutex
	cache = persistentCache[T]{&mutex, path}
	return
}

func (c persistentCache[T]) bucketFilepath(key string) (dir, path string) {
	hashedKey := hashKey(key)
	level1 := hashedKey[:2]
	level2 := hashedKey[2:4]
	dir = filepath.Join(c.path, level1, level2)
	path = filepath.Join(dir, hashedKey)
	return
}

func (c persistentCache[T]) Load(key string) (value T, ok bool, err error) {
	_, bucketPath := c.bucketFilepath(key)
	//fmt.Printf("Loading value from bucket: %s\n", bucketPath)
	c.mutex.Lock()
	content, err := os.ReadFile(bucketPath)
	c.mutex.Unlock()
	if os.IsNotExist(err) {
		err = nil
		return
	} else if err != nil {
		return
	}

	ok = true
	switch t := any(value).(type) {
	case string:
		value = any(string(content)).(T)
	default:
		log.Fatalf("Cannot Load value of type: %v ! Not supported yet.", t)
	}
	return
}

func (c persistentCache[T]) Store(key string, value T) (err error) {
	dir, path := c.bucketFilepath(key)
	c.mutex.Lock()
	// FIXME: always atempt to create dir
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return
	}
	//fmt.Printf("Storing value: %s in bucket: %s ...\n", value, bucket)
	switch t := any(value).(type) {
	case string:
		err = os.WriteFile(path, []byte(any(value).(string)), 0644)
	default:
		log.Fatalf("Cannot Store value of type: %v ! Not supported yet.", t)
	}
	c.mutex.Unlock()
	return
}

func hashKey(key string) (h string) {
	hBytes := sha256.Sum256([]byte(key))
	h = hex.EncodeToString(hBytes[:])
	return
}
