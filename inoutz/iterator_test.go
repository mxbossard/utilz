package inoutz

import (
	"strings"
	"testing"

	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitIterator_All_BigBuffer(t *testing.T) {
	s := strings.NewReader(ztring.LoremIpsumWords(10))
	sr := NewSplitIterator(s, ' ', 1000)
	require.NotNil(t, sr)
	k := 0
	errChan := make(chan error)
	for pos, data := range sr.All(errChan) {
		assert.Equal(t, ztring.SplitedLoremIpsumWords()[pos], string(data))
		k++
	}
	assert.Len(t, errChan, 0)
	assert.Equal(t, 10, k)
}

func TestSplitIterator_All_SmallBuffer(t *testing.T) {
	s := strings.NewReader(ztring.LoremIpsumWords(10))
	sr := NewSplitIterator(s, ' ', 25)
	require.NotNil(t, sr)
	k := 0
	errChan := make(chan error)
	for pos, data := range sr.All(errChan) {
		assert.Equal(t, ztring.SplitedLoremIpsumWords()[pos], string(data))
		k++
	}
	assert.Len(t, errChan, 0)
	assert.Equal(t, 10, k)
}

func TestSplitIterator_All_TinyBuffer(t *testing.T) {
	s := strings.NewReader(ztring.LoremIpsumWords(10))
	sr := NewSplitIterator(s, ' ', 15)
	require.NotNil(t, sr)
	k := 0
	errChan := make(chan error)
	for pos, data := range sr.All(errChan) {
		assert.Equal(t, ztring.SplitedLoremIpsumWords()[pos], string(data))
		k++
	}
	assert.Len(t, errChan, 0)
	assert.Equal(t, 10, k)
}
