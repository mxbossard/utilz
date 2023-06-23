package inout

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordingWriter(t *testing.T) {
	nested := strings.Builder{}
	rw := RecordingWriter{}
	rw.Nested = &nested

	expectedMsg1 := "foo"
	expectedMsg2 := "bar"
	expectedMsg3 := "baz"

	n, err := rw.Write([]byte(expectedMsg1))
	require.NoError(t, err)
	assert.Equal(t, len(expectedMsg1), n)
	assert.Equal(t, expectedMsg1, nested.String())
	assert.Equal(t, expectedMsg1, rw.String())

	n, err = rw.Write([]byte(expectedMsg2))
	require.NoError(t, err)
	assert.Equal(t, len(expectedMsg2), n)
	assert.Equal(t, expectedMsg1+expectedMsg2, nested.String())
	assert.Equal(t, expectedMsg1+expectedMsg2, rw.String())

	rw.Reset()
	n, err = rw.Write([]byte(expectedMsg3))
	require.NoError(t, err)
	assert.Equal(t, len(expectedMsg3), n)
	assert.Equal(t, expectedMsg1+expectedMsg2+expectedMsg3, nested.String())
	assert.Equal(t, expectedMsg3, rw.String())
}
