package cache

import (
	//"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mby.fr/utils/test"
)

func TestFileCacheStoreAndLoad(t *testing.T) {
	path, _ := test.BuildRandTempPath()
	defer os.RemoveAll(path)

	cache, err := NewPersistentCache[string](path)
	require.NoError(t, err)

	key := "test"
	value := "val"
	err = cache.Store(key, value)

	assert.NoError(t, err, "Store() should not return an error")

	res, ok, err := cache.Load(key)
	assert.True(t, ok, "Load() should return ok")
	assert.NoError(t, err, "Load() should not return an error")
	assert.Equal(t, value, res, "bad Load() return value ")
}

func TestFileCacheStoreEmptyValue(t *testing.T) {
	path, _ := test.BuildRandTempPath()
	defer os.RemoveAll(path)

	cache, err := NewPersistentCache[string](path)
	require.NoError(t, err)

	key := "test"
	value := ""
	err = cache.Store(key, value)

	require.NoError(t, err, "Store() should not return an error")

	res, ok, err := cache.Load(key)
	assert.True(t, ok, "Load() should return ok")
	assert.NoError(t, err, "Load() should not return an error")
	assert.Equal(t, value, res, "bad Load() return value ")
}

func TestFileCacheStoreSpecialChars(t *testing.T) {
	path, _ := test.BuildRandTempPath()
	defer os.RemoveAll(path)

	cache, err := NewPersistentCache[string](path)
	require.NoError(t, err)

	key := "test"
	value := `<>!$\n42?'foo"bar%baz`
	err = cache.Store(key, value)

	assert.NoError(t, err, "Store() should not return an error")

	res, ok, err := cache.Load(key)
	assert.True(t, ok, "Load() should return ok")
	assert.NoError(t, err, "Load() should not return an error")
	assert.Equal(t, value, res, "bad Load() return value ")
}

func TestFileCachePersistence(t *testing.T) {
	path, _ := test.BuildRandTempPath()
	defer os.RemoveAll(path)

	cache, err := NewPersistentCache[string](path)
	require.NoError(t, err)

	key := "test"
	value := "val"
	err = cache.Store(key, value)

	assert.DirExists(t, path, "FileCache dir should exists")
	assert.NoError(t, err, "Store() should not return an error")

	cache2, err := NewPersistentCache[string](path)
	require.NoError(t, err)
	res, ok, err := cache2.Load(key)
	assert.True(t, ok, "Load() should return ok")
	assert.NoError(t, err, "Load() should not return an error")
	assert.Equal(t, value, res, "bad Load() return value ")

	os.RemoveAll(path)

	cache3, err := NewPersistentCache[string](path)
	require.NoError(t, err)
	res, ok, err = cache3.Load(key)
	assert.False(t, ok, "Load() should not return ok")
	assert.NoError(t, err, "Load() should not return an error")
	assert.Equal(t, "", res, "Load() should return the empty string")
}

func TestFileCacheLoadNotExistingKey(t *testing.T) {
	path, _ := test.BuildRandTempPath()
	defer os.RemoveAll(path)

	cache, err := NewPersistentCache[string](path)
	require.NoError(t, err)

	key := "test"

	res, ok, err := cache.Load(key)
	assert.False(t, ok, "Load() should not return ok")
	assert.NoError(t, err, "Load() should not return an error")
	assert.Equal(t, "", res, "Load() should return the empty string")
}
