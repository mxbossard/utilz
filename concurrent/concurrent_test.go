package concurrent

import (
	//"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

var ()

func TestRun(t *testing.T) {
	expected := []string{"foo", "bar", "baz"}
	p := func(a string) (string, error) {
		return a, nil
	}

	wg, outputs, errors := Run[](p, expected...)

	assert.Empty(t, errors)
	//assert.Empty(t, outputs)
	assert.True(t, len(outputs) < len(expected))
	wg.Wait()
	assert.Empty(t, errors)
	assert.NotEmpty(t, outputs)
	assert.Len(t, outputs, len(expected))
}

func TestRunWaiting(t *testing.T) {
	expected := []string{"foo", "bar", "baz"}
	p := func(a string) (string, error) {
		return a, nil
	}

	outputs, errors := RunWaiting(p, expected...)

	assert.Empty(t, errors)
	assert.NotEmpty(t, outputs)
	assert.Len(t, outputs, len(expected))
}
