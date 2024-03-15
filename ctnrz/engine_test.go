package ctnrz

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
	run := Engine().Container().Run(testImage, "echo", expectedOut).Rm().Executer()
	var outBuff bytes.Buffer
	var errBuff bytes.Buffer
	run.SetOutputs(&outBuff, &errBuff)
	exitCode, err := run.BlockRun()

	assert.Equal(t, expectedOut+"\n", outBuff.String())
	assert.Empty(t, errBuff.String())
	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
}

func TestWaitRunWithEntrypoint(t *testing.T) {
	expectedOut := "foo"
	run := Engine().Container().Run(testImage, expectedOut).Entrypoint("echo").Rm().Executer()
	var outBuff bytes.Buffer
	var errBuff bytes.Buffer
	run.SetOutputs(&outBuff, &errBuff)
	exitCode, err := run.BlockRun()

	assert.Equal(t, expectedOut+"\n", outBuff.String())
	assert.Empty(t, errBuff.String())
	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
}

func TestWaitRunWithEnvArg(t *testing.T) {
	expectedOut := "foo"
	envArgs := make(map[string]string)
	envArgs["var"] = expectedOut
	run := Engine().Container().Run(testImage, "sh", "-c", "echo $var").AddEnvMap(envArgs).Rm().Executer()
	var outBuff bytes.Buffer
	var errBuff bytes.Buffer
	run.SetOutputs(&outBuff, &errBuff)
	exitCode, err := run.BlockRun()

	assert.Equal(t, expectedOut+"\n", outBuff.String())
	assert.Empty(t, errBuff.String())
	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
}
