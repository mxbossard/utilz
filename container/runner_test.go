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
	run := DockerRunner{Remove: true, Image: testImage, CmdArgs: []string{"echo", expectedOut}}
	var outBuff bytes.Buffer
	var errBuff bytes.Buffer
	err := run.Wait(&outBuff, &errBuff)

	assert.Equal(t, expectedOut+"\n", outBuff.String())
	assert.Empty(t, errBuff.String())
	require.NoError(t, err)
}

func TestWaitRunWithEntrypoint(t *testing.T) {
	expectedOut := "foo"
	run := DockerRunner{Remove: true, Image: testImage, Entrypoint: "echo", CmdArgs: []string{expectedOut}}
	var outBuff bytes.Buffer
	var errBuff bytes.Buffer
	err := run.Wait(&outBuff, &errBuff)

	assert.Equal(t, expectedOut+"\n", outBuff.String())
	assert.Empty(t, errBuff.String())
	require.NoError(t, err)
}

func TestWaitRunWithEnvArg(t *testing.T) {
	expectedOut := "foo"
	envArgs := make(map[string]string)
	envArgs["var"] = expectedOut
	run := DockerRunner{Remove: true, Image: testImage, EnvArgs: envArgs, CmdArgs: []string{"sh", "-c", "echo $var"}}
	var outBuff bytes.Buffer
	var errBuff bytes.Buffer
	err := run.Wait(&outBuff, &errBuff)

	assert.Equal(t, expectedOut+"\n", outBuff.String())
	assert.Empty(t, errBuff.String())
	require.NoError(t, err)
}
