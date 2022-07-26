package container

import (
	//"bytes"

	"io"
	"os/exec"

	"mby.fr/utils/errorz"
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
	errorsChan := make(chan error, 10)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errorsChan <- err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		errorsChan <- err
	}
	go inout.CopyChannelingErrors(stdout, stdOut, errorsChan)
	go inout.CopyChannelingErrors(stderr, stdErr, errorsChan)

	err = cmd.Start()
	if err != nil {
		errorsChan <- err
	}
	err = cmd.Wait()
	if err != nil {
		errorsChan <- err
	}

	// Aggregate all errors
	var errors errorz.Aggregated
	for {
		var err error
		// Use select to not block if no error in channel
		select {
		case err = <-errorsChan:
			errors.Add(err)
		default:
		}
		if err == nil {
			break
		}
	}

	return errors.Return()
}
