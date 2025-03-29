package collectionz

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

var (
	numArray = []int{1, 3, 4, 2, 0, 5, 8, 9, 7, 6}
	binArray = []int{1, 2, 4, 8, 16, 32, 0}
	strArray = []string{"bbb", "bba", "aaa", "aab", "ccc", "aac"}
)

func TestFilter(t *testing.T) {
	f1 := func(i int) bool { return i < 3 }
	col := Filter[int](&numArray, f1)
	assert.Equal(t, []int{1, 2, 0}, col)

	f2 := func(i int) bool { return i > 6 }
	col = Filter[int](&numArray, f2)
	assert.Equal(t, []int{8, 9, 7}, col)
}

func TestMap(t *testing.T) {
	multiply := func(i int) int { return i * 2 }
	col := Map[int, int](&numArray, multiply)
	assert.Equal(t, []int{2, 6, 8, 4, 0, 10, 16, 18, 14, 12}, col)
}

func TestReduce(t *testing.T) {
	summer := func(i, j int) int { return (i + j) }
	s := Reduce[int](&numArray, summer)
	assert.Equal(t, 45, s)
}

func TestContains(t *testing.T) {
	assert.True(t, Contains[int](&numArray, 3))
	assert.False(t, Contains[int](&numArray, 12))

	assert.True(t, Contains[string](&strArray, "aaa"))
	assert.False(t, Contains[string](&strArray, "aa"))
}

func TestKeepLeft(t *testing.T) {
	col := KeepLeft(&numArray, &binArray)
	assert.Equal(t, []int{3, 5, 9, 7, 6}, col)
}

func TestIntersect(t *testing.T) {
	col := Intersect(&numArray, &binArray)
	assert.Equal(t, []int{1, 4, 2, 0, 8}, col)
}

func TestDeduplicate(t *testing.T) {
	cat := append(numArray, binArray...)
	col := Deduplicate(&cat)
	expected := []int{1, 3, 4, 2, 0, 5, 8, 9, 7, 6, 16, 32}
	sort.Ints(expected)
	sort.Ints(col)
	assert.Equal(t, expected, col)
}
