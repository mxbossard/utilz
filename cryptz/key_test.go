package cryptz

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyHolder(t *testing.T) {
	kh := NewStringKeyHolder("foo", "bar")

	salt32, err := kh.Salt(32)
	assert.NoError(t, err)
	assert.Len(t, salt32, 32)
	fmt.Printf("salt32: %v", salt32)

	key32, err := kh.Key(32)
	assert.NoError(t, err)
	assert.Len(t, key32, 32)
	fmt.Printf("key32: %v", key32)

	key256, err := kh.Key(256)
	assert.NoError(t, err)
	assert.Len(t, key256, 256)
	fmt.Printf("key256: %v", key256)
}
