package errorz

import (
	"fmt"
	"testing"

	//"fmt"

	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

type otherTimeout struct {
	timeout
}

func (e otherTimeout) Error() string {
	return "otherTimeout"
}

func TestNewTimeout(t *testing.T) {
	err := Timeout(1, "foo")
	assert.NotNil(t, err)
	assert.True(t, IsTimeout(err))

	wrapped := fmt.Errorf("wrapping %w", err)
	assert.True(t, IsTimeout(wrapped))

	other := fmt.Errorf("other")
	assert.False(t, IsTimeout(other))

	otherTimeout := otherTimeout{timeout{Duration: 1, Message: "bar"}}
	assert.False(t, IsTimeout(otherTimeout))
}
