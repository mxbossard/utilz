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

type Provider string

const (
	CONTAINER_ENGINE_ENV_KEY = "CONTAINER_ENGINE"

	RUNNING   = Status("RUNNING")
	STOPPED   = Status("STOPPED")
	NOT_FOUND = Status("NOT_FOUND")

	PODMAN = Provider("PODMAN")
	DOCKER = Provider("DOCKER")
)

/*
	type common struct {
		container
	}
*/
type runner struct {
	container

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
	params := []string{c.binary, "run"}
	if c.name != "" {
		params = append(params, "--name", c.name)
	}
	if c.remove {
		params = append(params, "--rm")
	}
	if c.detach {
		params = append(params, "-d")
	}
	if c.entrypoint != "" {
		params = append(params, "--entrypoint", c.entrypoint)
	}
	if c.user != "" {
		params = append(params, "-u", c.user)
	}
	for _, arg := range c.volumes {
		params = append(params, "-v", arg)
	}
	for _, envArg := range c.envArgs {
		var envArg string = "-e=" + envArg
		params = append(params, envArg)
	}

	// podman specific
	if c.provider == PODMAN {
		// FIXME: always do this for podman ?
		params = append(params, "--userns", "keep-id")
	}

	params = append(params, c.image)
	// Add command args
	params = append(params, c.cmdAndArgs...)

	e := cmdz.Cmd(params...).ErrorOnFailure(true)
	if c.timeout != nil {
		e.Timeout(*c.timeout)
	}
	return e
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

func (c *runner) AddEnvs(envs ...string) *runner {
	c.envArgs = append(c.envArgs, envs...)
	return c
}

func (c *runner) AddEnvMap(envs map[string]string) *runner {
	for key, value := range envs {
		c.AddEnvs(key + "=" + value)
	}
	return c
}

func (c *runner) AddVolumes(vols ...string) *runner {
	c.volumes = append(c.volumes, vols...)
	return c
}

func (c *runner) Timeout(timeout time.Duration) *runner {
	c.timeout = &timeout
	return c
}

type executer struct {
	container

	cmdAndArgs []string

	interactive bool
	tty         bool
	detach      bool
	user        string
	envArgs     []string
}

func (c *executer) Executer() cmdz.Executer {
	params := []string{c.binary, "exec"}
	if c.interactive {
		params = append(params, "-i")
	}
	if c.tty {
		params = append(params, "-t")
	}
	if c.user != "" {
		params = append(params, "-u", c.user)
	}
	for _, envArg := range c.envArgs {
		var envArg string = "-e=" + envArg
		params = append(params, envArg)
	}
	params = append(params, c.name)
	params = append(params, c.cmdAndArgs...)

	e := cmdz.Cmd(params...).ErrorOnFailure(true)
	if c.timeout != nil {
		e.Timeout(*c.timeout)
	}
	return e
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

func (c *executer) AddEnvs(envs ...string) *executer {
	c.envArgs = append(c.envArgs, envs...)
	return c
}

func (c *executer) AddEnvMap(envs map[string]string) *executer {
	for key, value := range envs {
		c.AddEnvs(key + "=" + value)
	}
	return c
}

func (c *executer) Timeout(timeout time.Duration) *executer {
	c.timeout = &timeout
	return c
}

type starter struct {
	container
}

func (c *starter) Executer() cmdz.Executer {
	params := []string{c.binary, "start"}
	params = append(params, c.name)
	e := cmdz.Cmd(params...).ErrorOnFailure(true)
	if c.timeout != nil {
		e.Timeout(*c.timeout)
	}
	return e
}

func (c *starter) Timeout(timeout time.Duration) *starter {
	c.timeout = &timeout
	return c
}

type stopper struct {
	container
}

func (c *stopper) Executer() cmdz.Executer {
	params := []string{c.binary, "stop"}
	if c.timeout != nil {
		params = append(params, "--time", fmt.Sprintf("%d", int64(c.timeout.Round(time.Second).Seconds())))
	}
	params = append(params, c.name)
	e := cmdz.Cmd(params...).ErrorOnFailure(true)
	return e
}

func (c *stopper) Timeout(timeout time.Duration) *stopper {
	c.timeout = &timeout
	return c
}

type pser struct {
	engine

	timeout *time.Duration
}

func (c *pser) Executer() cmdz.Executer {
	params := []string{c.binary, "ps"}
	e := cmdz.Cmd(params...).ErrorOnFailure(true)
	if c.timeout != nil {
		e.Timeout(*c.timeout)
	}
	return e
}

func (c *pser) Timeout(timeout time.Duration) *pser {
	c.timeout = &timeout
	return c
}

type builder struct {
	engine

	buildCtx string
	tag      string

	noCache bool
	timeout *time.Duration
}

func (c *builder) Executer() cmdz.Executer {
	params := []string{c.binary, "build", "-t", c.tag}
	if c.noCache {
		params = append(params, "--no-cache")
	}
	params = append(params, c.buildCtx)
	e := cmdz.Cmd(params...).ErrorOnFailure(true)
	if c.timeout != nil {
		e.Timeout(*c.timeout)
	}
	return e
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
	engine

	image string

	timeout *time.Duration
}

func (c *puller) Executer() cmdz.Executer {
	params := []string{c.binary, "pull", c.image}
	e := cmdz.Cmd(params...).ErrorOnFailure(true)
	if c.timeout != nil {
		e.Timeout(*c.timeout)
	}
	return e
}

func (c *puller) Timeout(timeout time.Duration) *puller {
	c.timeout = &timeout
	return c
}

type pusher struct {
	engine

	image string

	timeout *time.Duration
}

func (c *pusher) Executer() cmdz.Executer {
	params := []string{c.binary, "push", c.image}
	e := cmdz.Cmd(params...).ErrorOnFailure(true)
	if c.timeout != nil {
		e.Timeout(*c.timeout)
	}
	return e
}

func (c *pusher) Timeout(timeout time.Duration) *pusher {
	c.timeout = &timeout
	return c
}

type tagger struct {
	engine

	image string
	tag   string

	timeout *time.Duration
}

func (c *tagger) Executer() cmdz.Executer {
	params := []string{c.binary, "tag", c.image}
	e := cmdz.Cmd(params...).ErrorOnFailure(true)
	if c.timeout != nil {
		e.Timeout(*c.timeout)
	}
	return e
}

func (c *tagger) Timeout(timeout time.Duration) *tagger {
	c.timeout = &timeout
	return c
}

type engine struct {
	provider Provider
	binary   string
}

func (e engine) Container(names ...string) *container {
	var name string
	if len(names) == 0 {
		uuid, err := utilz.ForgeUuid()
		if err != nil {
			panic(err)
		}
		name = uuid
	} else if len(names) == 1 {
		name = names[0]
	} else {
		panic("ctnrz.engine.Container() take 0 or 1 name as argument not more")
	}
	return &container{engine: e, name: name}
}

func (e engine) Ps() *pser {
	return &pser{}
}

func (e engine) Build(buildCtxDir, tag string) *builder {
	return &builder{buildCtx: buildCtxDir, tag: tag}
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
	panic("not implemented yet")
}

func (e engine) Exists(name string) bool {
	panic("not implemented yet")
}

func (e engine) IsRunning(name string) bool {
	panic("not implemented yet")
}

func (e engine) IsStopped(name string) bool {
	panic("not implemented yet")
}

type container struct {
	engine

	name    string
	timeout *time.Duration
}

func (c container) Run(image string, cmdAndArgs ...string) *runner {
	return &runner{container: c, image: image, cmdAndArgs: cmdAndArgs}
}

func (c container) Exec(cmdAndArgs ...string) *executer {
	return &executer{container: c, cmdAndArgs: cmdAndArgs}
}

func (c container) Stop() *stopper {
	return &stopper{container: c}
}

func Engine() engine {
	ok, binaryPath := selectContainerEngine()
	if !ok {
		err := fmt.Errorf("no container engine found in path. You must install podman or docker")
		panic(err)
	}
	var provider Provider
	if strings.Contains(binaryPath, "podman") {
		provider = PODMAN
	} else if strings.Contains(binaryPath, "docker") {
		provider = DOCKER
	}
	return engine{provider: provider, binary: binaryPath}
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
