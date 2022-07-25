package container

import (
	//"bytes"

	"io"
	"os/exec"

	"mby.fr/utils/inout"
)

var (
	binary = "docker"
)

type Runner struct {
	Name       string
	Remove     bool
	Entrypoint string
	EnvArgs    map[string]string
	Volumes    []string
	Image      string
	CmdArgs    []string
}

func (r Runner) Wait(stdOut io.Writer, stdErr io.Writer) (err error) {
	var runParams []string
	runParams = append(runParams, "run")

	if r.Name != "" {
		runParams = append(runParams, "--name", r.Name)
	}

	if r.Remove {
		runParams = append(runParams, "--rm")
	}

	if r.Entrypoint != "" {
		runParams = append(runParams, "--entrypoint", r.Entrypoint)
	}

	// Add volumes args
	for _, arg := range r.Volumes {
		runParams = append(runParams, "-v", arg)
	}

	// Add env args
	for argKey, argValue := range r.EnvArgs {
		var envArg string = "-e=" + argKey + "=" + argValue
		runParams = append(runParams, envArg)
	}

	runParams = append(runParams, r.Image)

	// Add command args
	runParams = append(runParams, r.CmdArgs...)

	cmd := exec.Command(binary, runParams...)

	// Manage // exec outputs
	errors := make(chan error, 10)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errors <- err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		errors <- err
	}
	go inout.CopyChannelingErrors(stdout, stdOut, errors)
	go inout.CopyChannelingErrors(stderr, stdErr, errors)

	err = cmd.Start()
	if err != nil {
		errors <- err
	}
	err = cmd.Wait()
	// Return first error from Wait() or from errors chan.
	if err == nil {
		// Use select to not block if no error in channel
		select {
		case err = <-errors:
		default:
		}
	}

	return
}
