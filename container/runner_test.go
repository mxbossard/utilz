package container

import (
	//"fmt"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testImage = "alpine:3.16"
)

func TestWaitRun(t *testing.T) {
	expectedOut := "foo"
	run := Runner{Image: testImage, CmdArgs: []string{"echo", expectedOut}}
	var outBuff bytes.Buffer
	var errBuff bytes.Buffer
	err := run.Wait(&outBuff, &errBuff)

	require.NoError(t, err)
	assert.Equal(t, expectedOut+"\n", outBuff.String())
	assert.Empty(t, errBuff.String())
}

func TestWaitRunWithEntrypoint(t *testing.T) {
	expectedOut := "foo"
	run := Runner{Image: testImage, Entrypoint: "echo", CmdArgs: []string{expectedOut}}
	var outBuff bytes.Buffer
	var errBuff bytes.Buffer
	err := run.Wait(&outBuff, &errBuff)

	require.NoError(t, err)
	assert.Equal(t, expectedOut+"\n", outBuff.String())
	assert.Empty(t, errBuff.String())
}

func TestWaitRunWithEnvArg(t *testing.T) {
	expectedOut := "foo"
	envArgs := make(map[string]string)
	envArgs["var"] = expectedOut
	run := Runner{Image: testImage, EnvArgs: envArgs, CmdArgs: []string{"sh", "-c", "echo $var"}}
	var outBuff bytes.Buffer
	var errBuff bytes.Buffer
	err := run.Wait(&outBuff, &errBuff)

	require.NoError(t, err)
	assert.Equal(t, expectedOut+"\n", outBuff.String())
	assert.Empty(t, errBuff.String())
}
