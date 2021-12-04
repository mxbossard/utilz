package cache

import (
	//"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"mby.fr/utils/test"
)

func TestFileCacheStoreAndLoad(t *testing.T) {
	path, _ := test.BuildRandTempPath()
	defer os.RemoveAll(path)

	cache, err := NewPersistentCache(path)

	key := "test"
	value := "val"
	err = cache.StoreString(key, value)

	assert.NoError(t, err, "StoreString() should not return an error")

	res, ok, err := cache.LoadString(key)
	assert.True(t, ok, "LoadString() should return ok")
	assert.NoError(t, err, "LoadString() should not return an error")
	assert.Equal(t, value, res, "bad LoadString() return value ")
}

func TestFileCacheStoreEmptyValue(t *testing.T) {
	path, _ := test.BuildRandTempPath()
	defer os.RemoveAll(path)

	cache, err := NewPersistentCache(path)

	key := "test"
	value := ""
	err = cache.StoreString(key, value)

	assert.NoError(t, err, "StoreString() should not return an error")

	res, ok, err := cache.LoadString(key)
	assert.True(t, ok, "LoadString() should return ok")
	assert.NoError(t, err, "LoadString() should not return an error")
	assert.Equal(t, value, res, "bad LoadString() return value ")
}

func TestFileCacheStoreSpecialChars(t *testing.T) {
	path, _ := test.BuildRandTempPath()
	defer os.RemoveAll(path)

	cache, err := NewPersistentCache(path)

	key := "test"
	value := `<>!$\n42?'foo"bar%baz`
	err = cache.StoreString(key, value)

	assert.NoError(t, err, "StoreString() should not return an error")

	res, ok, err := cache.LoadString(key)
	assert.True(t, ok, "LoadString() should return ok")
	assert.NoError(t, err, "LoadString() should not return an error")
	assert.Equal(t, value, res, "bad LoadString() return value ")
}

func TestFileCachePersistence(t *testing.T) {
	path, _ := test.BuildRandTempPath()
	defer os.RemoveAll(path)

	cache, err := NewPersistentCache(path)

	key := "test"
	value := "val"
	err = cache.StoreString(key, value)

	assert.DirExists(t, path, "FileCache dir should exists")
	assert.NoError(t, err, "StoreString() should not return an error")

	cache2, err := NewPersistentCache(path)
	res, ok, err := cache2.LoadString(key)
	assert.True(t, ok, "LoadString() should return ok")
	assert.NoError(t, err, "LoadString() should not return an error")
	assert.Equal(t, value, res, "bad LoadString() return value ")

	os.RemoveAll(path)

	cache3, err := NewPersistentCache(path)
	res, ok, err = cache3.LoadString(key)
	assert.False(t, ok, "LoadString() should not return ok")
	assert.NoError(t, err, "LoadString() should not return an error")
	assert.Equal(t, "", res, "LoadString() should return the empty string")
}

func TestFileCacheLoadNotExistingKey(t *testing.T) {
	path, _ := test.BuildRandTempPath()
	defer os.RemoveAll(path)

	cache, err := NewPersistentCache(path)

	key := "test"

	res, ok, err := cache.LoadString(key)
	assert.False(t, ok, "LoadString() should not return ok")
	assert.NoError(t, err, "LoadString() should not return an error")
	assert.Equal(t, "", res, "LoadString() should return the empty string")
}
