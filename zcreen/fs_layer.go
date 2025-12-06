package zcreen

import (
	"fmt"
	"io"
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

var (
	notifiersDirFilepathMatcher = regexp.MustCompile(`[/\\]` + notifiersDir)
	notifierFilepathMatcher     = regexp.MustCompile(`[/\\]` + notifiersDir + `[/\\]([^.]+)\.(\d+)\.(\d+)$`)
	sessionsDirFilepathMatcher  = regexp.MustCompile(`[/\\]` + sessionsDir)
	sessionDirFilepathMatcher   = regexp.MustCompile(`[/\\]` + sessionsDir + `[/\\](\d+)__([^/\\]+)`)
	printerDirFilepathMatcher   = regexp.MustCompile(`[/\\]` + printersDir + `[/\\](\d+)__([^/\\]+)`)
	printerFilepathMatcher      = regexp.MustCompile(`[/\\]` + printersDir + `[/\\](\d+)__([^/\\]+)[/\\]([^./\\]+)\.(\d+)\.(\d+)$`)
)

func forgeSessionsDirPath(zcreenPath string) string {
	return filepath.Join(zcreenPath, sessionsDir)
}

func forgeSessionDirPath(zcreenPath, sessionName string, sessionPriority int) string {
	sessionDirPath := forgeSessionsDirPath(zcreenPath)
	name := fmt.Sprintf("%012d__%s", sessionPriority, sessionName)
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
	name := fmt.Sprintf("%012d__%s", printerPriority, printerName)
	return filepath.Join(forgePrintersDirPath(zcreenPath, sessionName, sessionPriority), name)
}

func forgeUniqContextSuffix() string {
	pid := os.Getpid()
	timestamp := time.Now().UnixNano()
	suffix := fmt.Sprintf(".%021d.%d", timestamp, pid)
	return suffix
}

func forgeNotiferPrinterFilepathes(zcreenPath, sessionName string, sessionPriority int) (string, string, string) {
	dir := forgeNotifiersDirPath(zcreenPath, sessionName, sessionPriority)
	suffix := forgeUniqContextSuffix()
	outFilepath := filepath.Join(dir, fmt.Sprintf("out%s", suffix))
	errFilepath := filepath.Join(dir, fmt.Sprintf("err%s", suffix))
	closeFilepath := filepath.Join(dir, fmt.Sprintf("closed%s", suffix))
	return outFilepath, errFilepath, closeFilepath
}

func forgeSessionPrinterFilepathes(zcreenPath, sessionName string, sessionPriority int, printerName string, printerPriority int) (string, string, string) {
	dir := forgePrinterDirPath(zcreenPath, sessionName, sessionPriority, printerName, printerPriority)
	suffix := forgeUniqContextSuffix()
	outFilepath := filepath.Join(dir, fmt.Sprintf("out%s", suffix))
	errFilepath := filepath.Join(dir, fmt.Sprintf("err%s", suffix))
	closeFilepath := filepath.Join(dir, fmt.Sprintf("closed%s", suffix))
	return outFilepath, errFilepath, closeFilepath
}

func forgePrioNameKey(priority int, name string) string {
	return fmt.Sprintf("%012d-%s", priority, name)
}

type printerPart struct {
	outPath string
	// outCursor int64
	outFile *os.File
	errPath string
	// errCursor int64
	errFile *os.File
	closed  bool
}

func (p *printerPart) outputAt(outs printz.Outputs, outStart, errStart int64, buf []byte) (i, j int, err error) {
	// Write all bytes from out file
	eof := false
	for !eof {
		n, err := p.outFile.ReadAt(buf[0:], outStart)
		if err != nil && err != io.EOF {
			return -1, -1, err
		} else if err == io.EOF {
			eof = true
		}
		outs.Out().Write(buf[0:n])
		i += n
	}

	// Write all bytes from err file
	eof = false
	for !eof {
		n, err := p.errFile.ReadAt(buf[0:], errStart)
		if err != nil && err != io.EOF {
			return -1, -1, err
		} else if err == io.EOF {
			eof = true
		}
		outs.Err().Write(buf[0:n])
		j += n
	}

	return
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

	// fmt.Printf("will scan printerGroup FS: %s ...\n", g.path)
	zcreenFs := os.DirFS(g.path)
	fs.WalkDir(zcreenFs, ".", func(path string, d fs.DirEntry, err error) error {
		path = filepath.Join(g.path, path)
		// fmt.Printf("scanning printerGroup %s %d file: %s\n", g.name, g.priority, path)
		var qualifier, timestamp, pid string
		if printerFilepathMatcher.MatchString(path) {
			submatches := printerFilepathMatcher.FindStringSubmatch(path)
			// priority := submatches[1]
			// name := submatches[2]
			qualifier = submatches[3]
			timestamp = submatches[4]
			pid = submatches[5]
		} else if notifierFilepathMatcher.MatchString(path) {
			submatches := notifierFilepathMatcher.FindStringSubmatch(path)
			qualifier = submatches[1]
			timestamp = submatches[2]
			pid = submatches[3]
		} else {
			// fmt.Printf("path: %s do not match notifierFilepathMatcher nor printerFilepathMatcher\n", path)
		}
		if qualifier != "" {
			partKey := fmt.Sprintf("%s-%s", timestamp, pid)
			// fmt.Printf("Adding printerPart #%s ...\n", partKey)
			part, ok := g.partsByKey[partKey]
			if !ok && qualifier == "out" {
				outFilePath := path
				errFilePath := ztring.ReplaceLast(path, "out", "err")
				g.partsByKey[partKey] = buildPrinterPart(outFilePath, errFilePath)
			} else if ok && qualifier == "closed" {
				part.closed = true
			}
		}
		return err
	})

	return agg.Return()
}

type sessionGroup struct {
	path     string
	name     string
	priority int

	notifierParts      *printerGroup
	printersByPrioName map[string]*printerGroup
	pointer            string
	started, ended     bool
	deadline           *time.Time
	// endMsg             string
}

func (g *sessionGroup) scanFiles() error {
	// TODO: update printer groups
	agg := errorz.NewAgg()

	// fmt.Printf("will scan sessionGroup FS: %s ...\n", g.path)
	zcreenFs := os.DirFS(g.path)
	fs.WalkDir(zcreenFs, ".", func(path string, d fs.DirEntry, err error) error {
		path = filepath.Join(g.path, path)
		// fmt.Printf("scanning sessionGroup %s %d file: %s\n", g.name, g.priority, path)
		// if g.notifierParts == nil {
		isNotifiersDir := notifiersDirFilepathMatcher.MatchString(path)
		if isNotifiersDir {
			// Match global notifier dir
			// g.notifierParts = buildPrinterGroup(path, "__notifier", 0)

			// Do not walk notifiers dir
			return fs.SkipDir
		}
		// }

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
			key := forgePrioNameKey(priority, name)
			g.printersByPrioName[key] = group

			// Do not walk session dir
			return fs.SkipDir
		} else {
			// fmt.Printf("path: %s do not match printerDirFilepathMatcher\n", path)
		}
		return err
	})

	if g.notifierParts != nil {
		err := g.notifierParts.scanFiles()
		agg.Add(err)
	}

	for _, v := range g.printersByPrioName {
		err := v.scanFiles()
		agg.Add(err)
	}

	return agg.Return()
}

type zcreenGroup struct {
	path               string
	notifierParts      *printerGroup
	sessionsByPrioName map[string]*sessionGroup
	pointer            string
}

func (g *zcreenGroup) Outputer() *zcreenGroupOutputer {
	return &zcreenGroupOutputer{
		zg:               g,
		cursors:          make(map[*os.File]int64),
		outputedSessions: make(map[*sessionGroup]bool),
		outputedParts:    make(map[*printerPart]bool),
	}
}

func (g *zcreenGroup) scanFiles() error {
	// TODO: scan z.path and add missing sessionGroups, printerGroups and printerFiles.
	// TODO: update printers and sessions states
	agg := errorz.NewAgg()

	// fmt.Printf("will scan zcreenGroup FS: %s ...\n", g.path)
	zcreenFs := os.DirFS(g.path)
	fs.WalkDir(zcreenFs, ".", func(path string, d fs.DirEntry, err error) error {
		path = filepath.Join(g.path, path)
		// fmt.Printf("scanning zcreenGroup path: %s\n", path)
		//isNotifiersPrinterPart := notifierFilepathMatcher.MatchString(path)
		// if g.notifierParts == nil {
		isNotifiersDir := notifiersDirFilepathMatcher.MatchString(path)
		if isNotifiersDir {
			// Match global notifier dir
			// g.notifierParts = buildPrinterGroup(path, "__notifier", 0)

			// Do not walk notifiers dir
			return fs.SkipDir
		}
		// }

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
			key := forgePrioNameKey(priority, name)
			g.sessionsByPrioName[key] = group

			// Do not walk session dir
			return fs.SkipDir
		}
		return err
	})

	if g.notifierParts != nil {
		err := g.notifierParts.scanFiles()
		agg.Add(err)
	}

	for _, v := range g.sessionsByPrioName {
		err := v.scanFiles()
		agg.Add(err)
	}

	return agg.Return()
}

func buildPrinterPart(outFilepath, errFilepath string) *printerPart {
	outFile := filez.OpenReadOnlyOrPanic(outFilepath)
	errFile := filez.OpenReadOnlyOrPanic(errFilepath)
	return &printerPart{
		outPath: outFilepath,
		// outCursor: 0,
		outFile: outFile,
		errPath: errFilepath,
		// errCursor: 0,
		errFile: errFile,
		closed:  false,
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
		notifierParts:      buildPrinterGroup(filepath.Join(sessionDir, notifiersDir), "__notifier", 0),
		printersByPrioName: make(map[string]*printerGroup),
		pointer:            "",
		started:            false,
		ended:              false,
		deadline:           nil,
	}
}

func buildZcreenGroup(zcreenDir string) *zcreenGroup {
	return &zcreenGroup{
		path:               zcreenDir,
		notifierParts:      buildPrinterGroup(filepath.Join(zcreenDir, notifiersDir), "__notifier", 0),
		sessionsByPrioName: make(map[string]*sessionGroup),
		pointer:            "",
	}
}

// Output all parts of a printer group keeping read context in cursors & outputedParts map.
// If waitForClosed => wait for each printer to be closed before continuing outputing.
func outputsPrinterGroup(cursors *map[*os.File]int64, outputedParts *map[*printerPart]bool, outs printz.Outputs, pg *printerGroup, buf []byte, waitForClosed bool) (err error) {
	printedSomeStuff := false
	fmt.Printf("outputing printer group: %s ...\n", pg.name)
	for _, pp := range pg.partsByKey {
		if (*outputedParts)[pp] {
			// Do not output closed printers already outputed.
			continue
		}
		fmt.Printf("outputing printer out part: %s ...\n", pp.outPath)
		i, j, err := pp.outputAt(outs, (*cursors)[pp.outFile], (*cursors)[pp.errFile], buf)
		if err != nil {
			return err
		}

		if i > 0 || j > 0 {
			(*cursors)[pp.outFile] += int64(i)
			(*cursors)[pp.errFile] += int64(j)
			printedSomeStuff = true
		}

		if pp.closed {
			// Flag printer part as outputed
			(*outputedParts)[pp] = true
		} else if waitForClosed {
			// current part not closed => exit loop
			break
		}
	}

	if printedSomeStuff {
		err = outs.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

type zcreenGroupOutputer struct {
	// TODO
	zg *zcreenGroup
	//sg             *sessionGroup
	//printerPointer *printerGroup
	cursors          map[*os.File]int64
	outputedSessions map[*sessionGroup]bool
	// outputedPrinters map[*printerGroup]bool
	outputedParts map[*printerPart]bool
}

func (o *zcreenGroupOutputer) outputsGlobalPrinter(outs printz.Outputs, pg *printerGroup, buf []byte, waitForClosed bool) (err error) {
	return outputsPrinterGroup(&o.cursors, &o.outputedParts, outs, pg, buf, waitForClosed)
}

func (o *zcreenGroupOutputer) Sessions() (names []string) {
	for _, v := range o.zg.sessionsByPrioName {
		names = append(names, v.name)
	}
	return
}

func (o *zcreenGroupOutputer) SessionOutputer(sessionName string) *sessionGroupOutputer {
	var sg *sessionGroup
	for _, v := range o.zg.sessionsByPrioName {
		if v.name == sessionName {
			sg = v
			break
		}
	}
	if sg == nil {
		panic(fmt.Sprintf("no session: %s to output", sessionName))
	}
	return &sessionGroupOutputer{
		zgo:              o,
		sg:               sg,
		cursors:          make(map[*os.File]int64),
		outputedPrinters: make(map[*printerGroup]bool),
		outputedParts:    make(map[*printerPart]bool),
	}
}

type sessionGroupOutputer struct {
	zgo *zcreenGroupOutputer
	sg  *sessionGroup
	//printerPointer *printerGroup
	cursors map[*os.File]int64
	// outputedSessions map[*sessionGroup]bool
	outputedPrinters map[*printerGroup]bool
	outputedParts    map[*printerPart]bool
}

func (o *sessionGroupOutputer) HasNext() bool {
	// return true while session not outputed
	return !o.zgo.outputedSessions[o.sg]
}

func (o *sessionGroupOutputer) outputsSessionPrinter(outs printz.Outputs, pg *printerGroup, buf []byte, waitForClosed bool) (err error) {
	return outputsPrinterGroup(&o.cursors, &o.outputedParts, outs, pg, buf, waitForClosed)
}

func (o *sessionGroupOutputer) Outputs(outs printz.Outputs, sessionName string) (err error) {
	// TODO: implem session timeout
	buf := make([]byte, 1024)

	// Attempt to output global notifier before outputing session
	err = o.zgo.outputsGlobalPrinter(outs, o.zgo.zg.notifierParts, buf, false)
	if err != nil {
		return err
	}

	for _, pg := range o.sg.printersByPrioName {
		// Attempt to output session notifier before outputing a printer.
		err = o.zgo.outputsGlobalPrinter(outs, o.sg.notifierParts, buf, false)
		if err != nil {
			return err
		}

		if o.outputedPrinters[pg] {
			continue
		}

		// Output from elected printer group
		err = o.outputsSessionPrinter(outs, pg, buf, true)
		if err != nil {
			return err
		}

		if pg.closed {
			// Flag printer group as outputed
			o.outputedPrinters[pg] = true
		} else {
			// current part not closed => exit loop
			break
		}
	}

	// Attempt to output session notifier after all printer were closed.
	err = o.zgo.outputsGlobalPrinter(outs, o.sg.notifierParts, buf, false)
	if err != nil {
		return err
	}

	if o.sg.ended {
		o.zgo.outputedSessions[o.sg] = true
		// attempt to output global notifier after session was ended.
		err = o.zgo.outputsGlobalPrinter(outs, o.zgo.zg.notifierParts, buf, false)
		if err != nil {
			return err
		}
	}

	return nil
}
