package cmdz

import (
	"io"
	"log"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"mby.fr/utils/stringz"
)

var commandMock *CmdMock

// Vanilla exec.Cmd
type Vcmd = exec.Cmd

type CmdMock struct {
	t        *testing.T
	callback func(Vcmd) (int, io.Reader, io.Reader)
}

func (m CmdMock) Mock(c *exec.Cmd) int {
	mockedRc, stdout, stderr := m.callback(*c)
	outBuffer := strings.Builder{}
	errBuffer := strings.Builder{}
	var err error
	if stdout != nil {
		_, err = io.Copy(&outBuffer, stdout)
		assert.NoError(m.t, err, "cmdz mock cannot copy stdout !")
	}
	if stderr != nil {
		_, err = io.Copy(&errBuffer, stderr)
		assert.NoError(m.t, err, "cmdz mock cannot copy stderr !")
	}

	outSummary := stringz.SummaryRatio(outBuffer.String(), 128, .2)
	errSummary := stringz.SummaryRatio(errBuffer.String(), 128, .2)
	textCmd := strings.Join(c.Args, " ")
	log.Printf("Test: %s Mocked cmd execution: [%s] returned RC=%d STDOUT=[%s] STDERR=[%s]", m.t.Name(), textCmd, mockedRc, outSummary, errSummary)

	_, err = io.Copy(c.Stdout, strings.NewReader(outBuffer.String()))
	assert.NoError(m.t, err, "cmdz mock cannot copy stdout !")
	_, err = io.Copy(c.Stderr, strings.NewReader(errBuffer.String()))
	assert.NoError(m.t, err, "cmdz mock cannot copy stderr !")

	// FIXME how to mock ProcessState ?
	//e.Cmd.ProcessState = &os.ProcessState{}

	return mockedRc
}

func StartMock(t *testing.T, callback func(c Vcmd) (rc int, stdout, stderr io.Reader)) {
	//mockingCommand = callback
	commandMock = &CmdMock{t, callback}
}

func StartStringMock(t *testing.T, callback func(c Vcmd) (rc int, stdout, stderr string)) {
	m := func(c exec.Cmd) (int, io.Reader, io.Reader) {
		rc, stdout, stderr := callback(c)
		return rc, strings.NewReader(stdout), strings.NewReader(stderr)
	}
	StartMock(t, m)
}

func StartSimpleMock(t *testing.T, rc int, stdout, stderr string) {
	callback := func(c Vcmd) (int, string, string) {
		return rc, stdout, stderr
	}
	StartStringMock(t, callback)
}

func StopMock() {
	//mockingCommand = nil
	commandMock = nil
}

func Contains(c Vcmd, parts ...string) bool {
	textCmd := strings.Join(c.Args, " ")

	contains := true
	for _, part := range parts {
		contains = contains && strings.Contains(textCmd, part)
	}
	return contains
}
