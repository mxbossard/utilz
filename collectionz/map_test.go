package collectionz

import (
	//"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

var (
	mapA = map[int]string{1: "a", 3: "c", 4: "d", 2: "b"}
)

func TestCloneMap(t *testing.T) {
	// TODO
}

func TestKeys(t *testing.T) {
	keys := Keys(mapA)
	assert.ElementsMatch(t, []int{1, 3, 4, 2}, keys)
}

func TestOrderedKeys(t *testing.T) {
	keys := OrderedKeys(mapA)
	assert.Equal(t, []int{1, 2, 3, 4}, keys)
}

func TestValues(t *testing.T) {
	values := Values(mapA)
	assert.ElementsMatch(t, []string{"a", "c", "d", "b"}, values)
}

func TestOrderedValues(t *testing.T) {
	values := OrderedValues(mapA)
	assert.Equal(t, []string{"a", "b", "c", "d"}, values)
}
