package ctnrz

import (
	"fmt"
	"os"
	"strings"
	"time"

	"mby.fr/utils/cmdz"
	"mby.fr/utils/utilz"
)

type Status string

const (
	CONTAINER_ENGINE_ENV_KEY = "CONTAINER_ENGINE"

	RUNNING   = Status("RUNNING")
	STOPPED   = Status("STOPPED")
	NOT_FOUND = Status("NOT_FOUND")
)

type common struct {
	name    string
	timeout *time.Duration
}

type runner struct {
	common

	image      string
	cmdAndArgs []string

	interactive  bool
	tty          bool
	remove       bool
	detach       bool
	privileged   bool
	user         string
	userns       string
	cpuLimit     float32
	memLimitInMb int
	volumes      []string
	envArgs      []string
	entrypoint   string
}

func (c *runner) Executer() cmdz.Executer {

}

func (c *runner) Interactive() *runner {
	c.interactive = true
	return c
}

func (c *runner) Tty() *runner {
	c.tty = true
	return c
}

func (c *runner) Rm() *runner {
	c.remove = true
	return c
}

func (c *runner) Detach() *runner {
	c.detach = true
	return c
}

func (c *runner) Privileged() *runner {
	c.privileged = true
	return c
}

func (c *runner) User(user string) *runner {
	c.user = user
	return c
}

func (c *runner) Entrypoint(entrypoint string) *runner {
	c.entrypoint = entrypoint
	return c
}

func (c *runner) Envs(envs ...string) *runner {
	c.envArgs = append(c.envArgs, envs...)
	return c
}

func (c *runner) Volumes(vols ...string) *runner {
	c.volumes = append(c.volumes, vols...)
	return c
}

func (c *runner) Timeout(timeout time.Duration) *runner {
	c.timeout = &timeout
	return c
}

type executer struct {
	common

	cmdAndArgs []string

	interactive bool
	tty         bool
	detach      bool
	user        string
	envArgs     []string
}

func (c *executer) Executer() cmdz.Executer {

}

func (c *executer) Interactive() *executer {
	c.interactive = true
	return c
}

func (c *executer) Tty() *executer {
	c.tty = true
	return c
}

func (c *executer) Detach() *executer {
	c.detach = true
	return c
}

func (c *executer) User(user string) *executer {
	c.user = user
	return c
}

func (c *executer) Envs(envs ...string) *executer {
	c.envArgs = append(c.envArgs, envs...)
	return c
}

func (c *executer) Timeout(timeout time.Duration) *executer {
	c.timeout = &timeout
	return c
}

type stoper struct {
	common
}

func (c *stoper) Executer() cmdz.Executer {

}

func (c *stoper) Timeout(timeout time.Duration) *stoper {
	c.timeout = &timeout
	return c
}

type pser struct {
	timeout *time.Duration
}

func (c *pser) Executer() cmdz.Executer {

}

func (c *pser) Timeout(timeout time.Duration) *pser {
	c.timeout = &timeout
	return c
}

type builder struct {
	buildCtx string
	noCache  bool

	timeout *time.Duration
}

func (c *builder) Executer() cmdz.Executer {

}

func (c *builder) NoCache() *builder {
	c.noCache = true
	return c
}

func (c *builder) Timeout(timeout time.Duration) *builder {
	c.timeout = &timeout
	return c
}

type puller struct {
	image string

	timeout *time.Duration
}

func (c *puller) Executer() cmdz.Executer {

}

func (c *puller) Timeout(timeout time.Duration) *puller {
	c.timeout = &timeout
	return c
}

type pusher struct {
	image string

	timeout *time.Duration
}

func (c *pusher) Executer() cmdz.Executer {

}

func (c *pusher) Timeout(timeout time.Duration) *pusher {
	c.timeout = &timeout
	return c
}

type tagger struct {
	image string
	tag   string

	timeout *time.Duration
}

func (c *tagger) Executer() cmdz.Executer {

}

func (c *tagger) Timeout(timeout time.Duration) *tagger {
	c.timeout = &timeout
	return c
}

type engine struct {
	binary string
}

func (e engine) Container(name string) *container {
	return &container{engine: e, name: name}
}

func (e engine) Ps() *pser {
	return &pser{}
}

func (e engine) Build(buildCtxDir string) *builder {
	return &builder{buildCtx: buildCtxDir}
}

func (e engine) Pull(image string) *puller {
	return &puller{image: image}
}

func (e engine) Push(image string) *pusher {
	return &pusher{image: image}
}

func (e engine) Tagger(image, tag string) *tagger {
	return &tagger{image: image, tag: tag}
}

func (e engine) Status(name string) Status {
}

func (e engine) Exists(name string) bool {
}

func (e engine) IsRunning(name string) bool {
}

func (e engine) IsStopped(name string) bool {
}

type container struct {
	engine

	name string
}

func (c container) Run(image string, cmdAndArgs ...string) *runner {
	return &runner{common: common{name: c.name}, image: image, cmdAndArgs: cmdAndArgs}
}

func (c container) Exec(cmdAndArgs ...string) *executer {
	return &executer{common: common{name: c.name}, cmdAndArgs: cmdAndArgs}
}

func (c container) Stop() *stoper {
	return &stoper{common: common{name: c.name}}
}

func Engine() engine {
	ok, binaryPath := selectContainerEngine()
	if !ok {
		err := fmt.Errorf("no container engine found in path. You must install podman or docker")
		panic(err)
	}
	return engine{binary: binaryPath}
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
