package zcreen

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/mxbossard/utilz/errorz"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/printz"
	"github.com/mxbossard/utilz/ztring"
)

const (
	notifiersDir = "__notifiers"
	sessionsDir  = "__sessions"
	printersDir  = "__printers"
)

func forgeSessionsDirPath(zcreenPath string) string {
	return filepath.Join(zcreenPath, sessionsDir)
}

func forgeSessionDirPath(zcreenPath, sessionName string, sessionPriority int) string {
	sessionDirPath := forgeSessionsDirPath(zcreenPath)
	name := fmt.Sprintf("%d__%s", sessionPriority, sessionName)
	return filepath.Join(sessionDirPath, name)
}

func forgeNotifiersDirPath(zcreenPath, sessionName string, sessionPriority int) string {
	if sessionName == "" {
		return filepath.Join(zcreenPath, notifiersDir)
	}
	return filepath.Join(forgeSessionDirPath(zcreenPath, sessionName, sessionPriority), notifiersDir)
}

func forgePrintersDirPath(zcreenPath, sessionName string, sessionPriority int) string {
	return filepath.Join(forgeSessionDirPath(zcreenPath, sessionName, sessionPriority), printersDir)
}

func forgePrinterDirPath(zcreenPath, sessionName string, sessionPriority int, printerName string, printerPriority int) string {
	name := fmt.Sprintf("%d__%s", printerPriority, printerName)
	return filepath.Join(forgeSessionDirPath(zcreenPath, sessionName, sessionPriority), name)
}

func forgeUniqContextSuffix() string {
	pid := os.Getpid()
	timestamp := time.Now().UnixNano()
	suffix := fmt.Sprintf(".%021d.%d", timestamp, pid)
	return suffix
}

func forgeNotiferPrinterFilepathes(zcreenPath, sessionName string, sessionPriority int) (string, string) {
	dir := forgeNotifiersDirPath(zcreenPath, sessionName, sessionPriority)
	suffix := forgeUniqContextSuffix()
	outFilepath := filepath.Join(dir, fmt.Sprintf("out%s", suffix))
	errFilepath := filepath.Join(dir, fmt.Sprintf("err%s", suffix))
	return outFilepath, errFilepath
}

func forgeSessionPrinterFilepathes(zcreenPath, sessionName string, sessionPriority int, printerName string, printerPriority int) (string, string) {
	dir := forgePrinterDirPath(zcreenPath, sessionName, sessionPriority, printerName, printerPriority)
	suffix := forgeUniqContextSuffix()
	outFilepath := filepath.Join(dir, fmt.Sprintf("out%s", suffix))
	errFilepath := filepath.Join(dir, fmt.Sprintf("err%s", suffix))
	return outFilepath, errFilepath
}

type printerPart struct {
	outPath   string
	outCursor int64
	outFile   *os.File
	errPath   string
	errCursor int64
	errFile   *os.File
	closed    bool
}

type printerGroup struct {
	path     string
	name     string
	priority int

	partsByKey map[string]*printerPart
	pointer    string
	closed     bool
	// closeMsg string
}

func (g *printerGroup) scanFiles() error {
	// TODO: update printer parts
	agg := errorz.NewAgg()

	zcreenFs := os.DirFS(g.path)
	fs.WalkDir(zcreenFs, ".", func(path string, d fs.DirEntry, err error) error {
		if printerFilepathMatcher.MatchString(path) {
			submatches := printerFilepathMatcher.FindStringSubmatch(path)
			// priority := submatches[1]
			// name := submatches[2]
			qualifier := submatches[3]
			timestamp := submatches[4]
			pid := submatches[5]

			partKey := fmt.Sprintf("%s-%s", timestamp, pid)
			if _, ok := g.partsByKey[partKey]; qualifier == "out" && !ok {
				outFilePath := path
				errFilePath := ztring.ReplaceLast(path, "out", "err")
				g.partsByKey[partKey] = buildPrinterPart(outFilePath, errFilePath)
			}
		}
		return nil
	})

	return agg.Return()
}

type sessionGroup struct {
	path     string
	name     string
	priority int

	notifierParts      *printerGroup
	printersByPrioName map[int]map[string]*printerGroup
	currentPriority    int
	currentName        string
	started, ended     bool
	deadline           *time.Time
	// endMsg             string
}

func (g *sessionGroup) scanFiles() error {
	// TODO: update printer groups
	agg := errorz.NewAgg()

	zcreenFs := os.DirFS(g.path)
	fs.WalkDir(zcreenFs, ".", func(path string, d fs.DirEntry, err error) error {
		if g.notifierParts == nil {
			isNotifiersDir := notifiersDirFilepathMatcher.MatchString(path)
			if isNotifiersDir {
				// Match global notifier dir
				g.notifierParts = buildPrinterGroup(path, "__notifier", 0)

				// Do not walk notifiers dir
				return fs.SkipDir
			}
		}

		isPrinterDir := printerDirFilepathMatcher.MatchString(path)
		if isPrinterDir {
			// Match a printer dir
			submatches := printerDirFilepathMatcher.FindStringSubmatch(path)
			priority, err := strconv.Atoi(submatches[1])
			if err != nil {
				panic(err)
			}
			name := submatches[2]
			group := buildPrinterGroup(path, name, priority)
			g.printersByPrioName[priority][name] = group

			// Do not walk session dir
			return fs.SkipDir
		}
		return nil
	})

	if g.notifierParts != nil {
		err := g.notifierParts.scanFiles()
		agg.Add(err)
	}

	for _, v := range g.printersByPrioName {
		for _, w := range v {
			err := w.scanFiles()
			agg.Add(err)
		}
	}

	return agg.Return()
}

type zcreenGroup struct {
	path               string
	notifierParts      *printerGroup
	sessionsByPrioName map[int]map[string]*sessionGroup
	currentPriority    int
	currentName        string
}

var (
	notifiersDirFilepathMatcher = regexp.MustCompile(`[/\\]` + notifiersDir)
	notifiersFilepathMatcher    = regexp.MustCompile(`[/\\]` + notifiersDir + `[/\\]([^.]+)\.(\d+)\\.(\d+)$`)
	sessionsDirFilepathMatcher  = regexp.MustCompile(`[/\\]` + sessionsDir)
	sessionDirFilepathMatcher   = regexp.MustCompile(`[/\\]` + sessionsDir + `[/\\](\d+)__([^/\\]+)`)
	printerDirFilepathMatcher   = regexp.MustCompile(`[/\\]` + printersDir + `[/\\](\d+)__([^/\\]+)`)
	printerFilepathMatcher      = regexp.MustCompile(`[/\\]` + printersDir + `[/\\](\d+)__([^/\\]+)[/\\]([^./\\]+)\.(\d+)\.(\d+)$`)
)

func (g *zcreenGroup) scanFiles() error {
	// TODO: scan z.path and add missing sessionGroups, printerGroups and printerFiles.
	// TODO: update printers and sessions states
	agg := errorz.NewAgg()

	zcreenFs := os.DirFS(g.path)
	fs.WalkDir(zcreenFs, ".", func(path string, d fs.DirEntry, err error) error {
		//isNotifiersPrinterPart := notifierFilepathMatcher.MatchString(path)
		if g.notifierParts == nil {
			isNotifiersDir := notifiersDirFilepathMatcher.MatchString(path)
			if isNotifiersDir {
				// Match global notifier dir
				g.notifierParts = buildPrinterGroup(path, "__notifier", 0)

				// Do not walk notifiers dir
				return fs.SkipDir
			}
		}

		isSessionDir := sessionDirFilepathMatcher.MatchString(path)
		if isSessionDir {
			// Match sessions dir
			submatches := sessionDirFilepathMatcher.FindStringSubmatch(path)
			priority, err := strconv.Atoi(submatches[1])
			if err != nil {
				panic(err)
			}
			name := submatches[2]
			group := buildSessionGroup(path, name, priority)
			g.sessionsByPrioName[priority][name] = group

			// Do not walk session dir
			return fs.SkipDir
		}
		return nil
	})

	if g.notifierParts != nil {
		err := g.notifierParts.scanFiles()
		agg.Add(err)
	}

	for _, v := range g.sessionsByPrioName {
		for _, w := range v {
			err := w.scanFiles()
			agg.Add(err)
		}
	}

	return agg.Return()
}

func buildPrinterPart(outFilepath, errFilepath string) *printerPart {
	outFile := filez.OpenOrPanic(outFilepath)
	errFile := filez.OpenOrPanic(errFilepath)
	return &printerPart{
		outPath:   outFilepath,
		outCursor: 0,
		outFile:   outFile,
		errPath:   errFilepath,
		errCursor: 0,
		errFile:   errFile,
		closed:    false,
	}
}

func buildPrinterGroup(printerDir, printerName string, printerPriority int) *printerGroup {
	return &printerGroup{
		path:       printerDir,
		name:       printerName,
		priority:   printerPriority,
		partsByKey: make(map[string]*printerPart),
		pointer:    "",
		closed:     false,
	}
}

func buildSessionGroup(sessionDir, sessionName string, sessionPriority int) *sessionGroup {
	return &sessionGroup{
		path:               sessionDir,
		name:               sessionName,
		priority:           sessionPriority,
		printersByPrioName: make(map[int]map[string]*printerGroup),
		currentPriority:    0,
		currentName:        "",
		started:            false,
		ended:              false,
		deadline:           nil,
	}
}

func buildZcreenGroup(zcreenDir string) *zcreenGroup {
	return &zcreenGroup{
		path:               zcreenDir,
		sessionsByPrioName: make(map[int]map[string]*sessionGroup),
		currentPriority:    0,
		currentName:        "",
	}
}

func buildFsOutputs(outFilepath, errFilepath string) (printz.Outputs, *os.File, *os.File) {
	filez.TouchMkdirAllOrPanic(outFilepath)
	filez.TouchMkdirAllOrPanic(errFilepath)
	outFile := filez.OpenOrPanic(outFilepath)
	errFile := filez.OpenOrPanic(errFilepath)
	outputs := printz.NewOutputs(outFile, errFile)
	return outputs, outFile, errFile
}

func buildFsPrinter(name string, priority int, outFilepath, errFilepath string, opened bool) *printer {
	outputs, outFile, errFile := buildFsOutputs(outFilepath, errFilepath)
	prtr := printz.New(outputs)
	closingPrtr := printz.Closing(prtr)
	p := &printer{
		ClosingPrinter: closingPrtr,
		name:           name,
		tmpOut:         outFile,
		tmpErr:         errFile,
		open:           opened,
		priorityOrder:  priority,
	}
	return p
}

func buildNotifierPrinter(zcreenPath, sessionName string, sessionPriority int, opened bool) *printer {
	notifierOutFilepath, notifierErrFilepath := forgeNotiferPrinterFilepathes(zcreenPath, sessionName, sessionPriority)
	return buildFsPrinter("__notifier", 0, notifierOutFilepath, notifierErrFilepath, opened)
}

func buildSessionPrinter(zcreenPath, sessionName string, sessionPriority int, printerName string, printerPriority int, opened bool) *printer {
	printerOutFilepath, printerErrFilepath := forgeSessionPrinterFilepathes(zcreenPath, sessionName, sessionPriority, printerName, printerPriority)
	return buildFsPrinter(printerName, printerPriority, printerOutFilepath, printerErrFilepath, opened)
}
