package container

import (
	//"bytes"

	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"mby.fr/utils/cmdz"
	"mby.fr/utils/errorz"
	"mby.fr/utils/inout"
	"mby.fr/utils/utilz"
)

const (
	CONTAINER_ENGINE_ENV_KEY = "CONTAINER_ENGINE"
)

var (
	dockerBinary = "docker"
)

type Runner interface {
	Wait(io.Writer, io.Writer) error
	//Async(io.Writer, io.Writer) error
}

type DockerRunner struct {
	Name       string
	Remove     bool
	Detach     bool
	Entrypoint string
	EnvArgs    map[string]string
	Volumes    []string
	Image      string
	CmdArgs    []string
}

func (r DockerRunner) Wait(stdOut io.Writer, stdErr io.Writer) (err error) {
	var runParams []string
	runParams = append(runParams, "run")

	if r.Name != "" {
		runParams = append(runParams, "--name", r.Name)
	}

	if r.Remove {
		runParams = append(runParams, "--rm")
	}

	if r.Detach {
		runParams = append(runParams, "-d")
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

	cmd := exec.Command(dockerBinary, runParams...)

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
	var errors = errorz.ConsumedAggregated(errorsChan)
	return errors.Return()
}

func (r DockerRunner) Async(stdOut io.Writer, stdErr io.Writer) {
	go func() {
		err2 := r.Wait(stdOut, stdErr)
		if err2 != nil {
			log.Fatal(err2)
		}
	}()
}

type LifeCycleBuilder interface {
	StopExecuter() cmdz.Executer
	StartExecuter() cmdz.Executer
	RmExecuter() cmdz.Executer
}

type RunBuilder interface {
	RunExecuter() cmdz.Executer
}

type ExecBuilder interface {
	ExecExecuter() cmdz.Executer
}

type LifeCycleConfig struct {
	Name    string
	Timeout time.Duration
}

type podmanLifeCycleBuilder struct {
	LifeCycleConfig
	binary string
}

func (r podmanLifeCycleBuilder) StopExecuter() cmdz.Executer {
	runParams := []string{r.binary, "stop", r.Name}
	cmd := cmdz.Cmd(runParams...).ErrorOnFailure(true)
	return cmd
}

func (r podmanLifeCycleBuilder) StartExecuter() cmdz.Executer {
	runParams := []string{r.binary, "start", r.Name}
	cmd := cmdz.Cmd(runParams...).ErrorOnFailure(true)
	return cmd
}

func (r podmanLifeCycleBuilder) RmExecuter() cmdz.Executer {
	runParams := []string{r.binary, "rm", "-f", r.Name}
	cmd := cmdz.Cmd(runParams...).ErrorOnFailure(true)
	return cmd
}

type dockerLifeCycleBuilder struct {
	podmanLifeCycleBuilder
}

type ExecConfig struct {
	Name        string
	Interactive bool
	Tty         bool
	User        string
	EnvArgs     map[string]string
	CmdAndArgs  []string

	Timeout time.Duration
}

func buildCommonExecParams(r ExecConfig) []string {
	var params []string

	if r.Interactive {
		params = append(params, "-i")
	}

	if r.Tty {
		params = append(params, "-t")
	}

	if r.User != "" {
		params = append(params, "-u", r.User)
	}

	// Add env args
	for argKey, argValue := range r.EnvArgs {
		var envArg string = "-e=" + argKey + "=" + argValue
		params = append(params, envArg)
	}

	params = append(params, r.Name)

	// Add command args
	params = append(params, r.CmdAndArgs...)

	return params
}

type podmanExecBuilder struct {
	ExecConfig
	binary string
}

func (r *podmanExecBuilder) ExecExecuter() cmdz.Executer {
	execParams := []string{r.binary, "exec"}
	execParams = append(execParams, buildCommonExecParams(r.ExecConfig)...)

	cmd := cmdz.Cmd(execParams...).ErrorOnFailure(true)
	return cmd
}

type dockerExecBuilder struct {
	podmanExecBuilder
}

type RunConfig struct {
	Name         string
	Interactive  bool
	Tty          bool
	Remove       bool
	Detach       bool
	Privileged   bool
	User         string
	Userns       string
	CpuLimit     float32
	MemLimitInMb int
	Volumes      []string
	EnvArgs      map[string]string
	Image        string
	Entrypoint   string
	CmdArgs      []string
}

func completeRunConfig(cfg *RunConfig) {
	if cfg.Name == "" {
		uuid, err := utilz.ForgeUuid()
		if err != nil {
			panic(err)
		}
		cfg.Name = uuid
	}
}

func buildCommonRunParams(r RunConfig) []string {
	var runParams []string

	if r.Name != "" {
		runParams = append(runParams, "--name", r.Name)
	}

	if r.Remove {
		runParams = append(runParams, "--rm")
	}

	if r.Detach {
		runParams = append(runParams, "-d")
	}

	if r.Entrypoint != "" {
		runParams = append(runParams, "--entrypoint", r.Entrypoint)
	}

	if r.User != "" {
		runParams = append(runParams, "-u", r.User)
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

	return runParams
}

type podmanRunBuilder struct {
	RunConfig
	binary string
}

func (r *podmanRunBuilder) RunExecuter() cmdz.Executer {
	completeRunConfig(&r.RunConfig)
	runParams := []string{r.binary, "run"}

	if r.Userns != "" {
		runParams = append(runParams, "--userns", r.Userns)
	}

	runParams = append(runParams, buildCommonRunParams(r.RunConfig)...)

	cmd := cmdz.Cmd(runParams...).ErrorOnFailure(true)
	return cmd
}

type dockerRunBuilder struct {
	podmanRunBuilder
}

func (r *dockerRunBuilder) RunExecuter() cmdz.Executer {
	completeRunConfig(&r.RunConfig)
	runParams := []string{r.binary, "run"}
	runParams = append(runParams, buildCommonRunParams(r.RunConfig)...)

	cmd := cmdz.Cmd(runParams...).ErrorOnFailure(true)
	return cmd
}

func PodmanRunBuilder(binary string, cfg RunConfig) *podmanRunBuilder {
	r := podmanRunBuilder{
		RunConfig: cfg,
		binary:    binary,
	}
	return &r
}

func DockerRunBuilder(binary string, cfg RunConfig) *dockerRunBuilder {
	r := dockerRunBuilder{
		podmanRunBuilder: podmanRunBuilder{
			RunConfig: cfg,
			binary:    binary,
		},
	}
	return &r
}

func NewRunBuilder(cfg RunConfig) RunBuilder {
	ok, binaryPath := selectContainerEngine()
	if !ok {
		err := fmt.Errorf("no container engine found in path. You must install podman or docker")
		panic(err)
	}
	if strings.Contains(binaryPath, "podman") {
		return PodmanRunBuilder(binaryPath, cfg)
	} else if strings.Contains(binaryPath, "docker") {
		return DockerRunBuilder(binaryPath, cfg)
	}
	err := fmt.Errorf("cannot determine which container engine is: %s", binaryPath)
	panic(err)
}

func PodmanExecBuilder(binary string, cfg ExecConfig) *podmanExecBuilder {
	r := podmanExecBuilder{
		ExecConfig: cfg,
		binary:     binary,
	}
	return &r
}

func DockerExecBuilder(binary string, cfg ExecConfig) *dockerExecBuilder {
	r := dockerExecBuilder{
		podmanExecBuilder: podmanExecBuilder{
			ExecConfig: cfg,
			binary:     binary,
		},
	}
	return &r
}

func NewExecBuilder(cfg ExecConfig) ExecBuilder {
	ok, binaryPath := selectContainerEngine()
	if !ok {
		err := fmt.Errorf("no container engine found in path. You must install podman or docker")
		panic(err)
	}
	if strings.Contains(binaryPath, "podman") {
		return PodmanExecBuilder(binaryPath, cfg)
	} else if strings.Contains(binaryPath, "docker") {
		return DockerExecBuilder(binaryPath, cfg)
	}
	err := fmt.Errorf("cannot determine which container engine is: %s", binaryPath)
	panic(err)
}

func PodmanLifeCycleBuilder(binary string, cfg LifeCycleConfig) podmanLifeCycleBuilder {
	r := podmanLifeCycleBuilder{
		LifeCycleConfig: cfg,
		binary:          binary,
	}
	return r
}

func DockerLifeCycleBuilder(binary string, cfg LifeCycleConfig) dockerLifeCycleBuilder {
	r := dockerLifeCycleBuilder{
		podmanLifeCycleBuilder: podmanLifeCycleBuilder{
			LifeCycleConfig: cfg,
			binary:          binary,
		},
	}
	return r
}

func NewLifeCycleBuilder(cfg LifeCycleConfig) LifeCycleBuilder {
	ok, binaryPath := selectContainerEngine()
	if !ok {
		err := fmt.Errorf("no container engine found in path. You must install podman or docker")
		panic(err)
	}
	if strings.Contains(binaryPath, "podman") {
		return PodmanLifeCycleBuilder(binaryPath, cfg)
	} else if strings.Contains(binaryPath, "docker") {
		return DockerLifeCycleBuilder(binaryPath, cfg)
	}
	err := fmt.Errorf("cannot determine which container engine is: %s", binaryPath)
	panic(err)
}

func selectContainerEngine() (ok bool, binaryPath string) {
	// Check presence of podman & docker binary
	// Prefer to use podman

	// Check env var for a preference
	ok, binaryPath = utilz.EnvValue(CONTAINER_ENGINE_ENV_KEY)
	if ok {
		return
	}

	// Search PATH for an existing binary
	pCmd := cmdz.Sh("which", "podman").ErrorOnFailure(false).AddEnviron(os.Environ()...)
	pCode, err := pCmd.BlockRun()
	if err != nil {
		panic(err)
	}
	if pCode == 0 {
		ok = true
		binaryPath = strings.TrimSpace(pCmd.StdoutRecord())
		return
	}

	dCmd := cmdz.Sh("which", "docker").ErrorOnFailure(false).AddEnviron(os.Environ()...)
	dCode, err := dCmd.BlockRun()
	if err != nil {
		panic(err)
	}
	if dCode == 0 {
		ok = true
		binaryPath = strings.TrimSpace(dCmd.StdoutRecord())
		return
	}

	return
}
