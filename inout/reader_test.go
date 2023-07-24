package inout

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordingReader(t *testing.T) {
	expectedMsg1 := "foo"
	nested := strings.NewReader(expectedMsg1)
	rr := RecordingReader{}
	rr.Nested = nested

	buffer := make([]byte, len(expectedMsg1)+5)
	n, err := rr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expectedMsg1), n)
	assert.Equal(t, expectedMsg1, string(buffer[0:n]))
	assert.Equal(t, expectedMsg1, rr.String())
}
