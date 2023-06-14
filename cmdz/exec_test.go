package cmdz

import (
	//"fmt"
	//"io"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddArgs(t *testing.T) {
	e := Execution("echo")
	assert.Equal(t, []string{"echo"}, e.Args)
	e.AddArgs("foo")
	assert.Equal(t, []string{"echo", "foo"}, e.Args)
	e.AddArgs("bar", "baz")
	assert.Equal(t, []string{"echo", "foo", "bar", "baz"}, e.Args)
}

func TestRunBlocking(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg := "foobar"
	e := Execution(echoBinary, echoArg)

	log.Printf("exec: %v", e)

	stdout := strings.Builder{}
	stderr := strings.Builder{}

	rc, err := e.RunBlocking(&stdout, &stderr)
	require.NoError(t, err, "should not error")
	assert.Equal(t, rc, 0)
	sout := stdout.String()
	serr := stderr.String()
	assert.Equal(t, echoArg+"\n", sout)
	assert.Equal(t, "", serr)
}
