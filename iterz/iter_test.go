package iterz

import (
	"cmp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMerge_Int(t *testing.T) {
	it1 := BuildIntSeq(1, 2, 5, 15, 12)
	it2 := BuildIntSeq(2, 3, 13, 7)
	it3 := BuildIntSeq(1, 4, 6, 14, 11, 8, 9)
	it4 := BuildIntSeq()
	expected := []int{1, 1, 2, 2, 3, 4, 5, 6, 13, 7, 14, 11, 8, 9, 15, 12}

	intCompare := func(a, b int) int {
		return cmp.Compare(a, b)
	}

	m := Merge(intCompare, it3, it1, it4, it2)
	var actual []int
	for i := range m {
		actual = append(actual, i)
	}

	assert.Equal(t, expected, actual)
}

func TestMerge_String(t *testing.T) {
	it1 := BuildStringSeq("1a", "2a", "5a", "9a", "3a")
	it2 := BuildStringSeq("2b", "3b", "7b", "6b")
	it3 := BuildStringSeq("1c", "4c", "6c", "9c", "8c", "5c", "6c")
	expected := []string{"1c", "1a", "2a", "2b", "3b", "4c", "5a", "6c", "7b", "6b", "9c", "8c", "5c", "6c", "9a", "3a"}

	stringCompare := func(a, b string) int {
		return cmp.Compare(a[0], b[0])
	}

	m := Merge(stringCompare, it3, it2, it1)
	var actual []string
	for i := range m {
		actual = append(actual, i)
	}

	assert.Equal(t, expected, actual)
}

func TestMerge2_Strings(t *testing.T) {
	itA1 := BuildStringSeq("1a", "2a", "5a", "9a", "3a")
	itA2 := BuildStringSeq("2b", "3b", "7b", "6b")
	itA := Link(itA1, itA2)
	itB1 := BuildStringSeq("1c", "4c", "6c", "9c", "8c", "5c", "6c")
	itB2 := BuildStringSeq("3d", "4d", "7d", "8d", "2d")
	itB := Link(itB1, itB2)
	itC := func(yield func(a, b string) bool) {
	}
	expectedK := []string{"1c", "1a", "2a", "4c", "5a", "6c", "9c", "8c", "9a"}
	expectedV := []string{"3d", "2b", "3b", "4d", "7b", "7d", "8d", "2d", "6b"}

	stringCompare := func(a1, a2, b1, b2 string) int {
		return cmp.Compare(a1[0], a2[0])
	}
	m := Merge2(stringCompare, itB, itC, itA)
	var actualK []string
	var actualV []string
	for k, v := range m {
		actualK = append(actualK, k)
		actualV = append(actualV, v)
	}

	assert.Equal(t, expectedK, actualK)
	assert.Equal(t, expectedV, actualV)
}
