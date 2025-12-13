package zcreen

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/mxbossard/utilz/collectionz"
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
	for !eof && p.outFile != nil {
		n, err := p.outFile.ReadAt(buf[0:], outStart)
		if err != nil && err != io.EOF {
			return -1, -1, fmt.Errorf("error ReadAt %d outFile: %w", outStart, err)
		} else if err == io.EOF {
			eof = true
		}
		outs.Out().Write(buf[0:n])
		i += n
	}

	// Write all bytes from err file
	eof = false
	for !eof && p.errFile != nil {
		n, err := p.errFile.ReadAt(buf[0:], errStart)
		if err != nil && err != io.EOF {
			return -1, -1, fmt.Errorf("error ReadAt %d outFile: %w", errStart, err)
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

	partsByKey  *map[string]*printerPart
	pointer     string
	closed      bool
	outputedOut int64
	outputedErr int64
	// closeMsg string
}

func (g *printerGroup) scanFiles() (updated bool, err error) {
	// TODO: update printer parts
	agg := errorz.NewAgg()

	// fmt.Printf("will scan printerGroup FS: %s ...\n", g.path)
	zcreenFs := os.DirFS(g.path)
	var scannedPrinterPartsKeys []string
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
			part, ok := (*g.partsByKey)[partKey]
			if ok {
				scannedPrinterPartsKeys = append(scannedPrinterPartsKeys, partKey)
			} else if !ok && qualifier == "out" {
				outFilePath := path
				errFilePath := ztring.ReplaceLast(path, "out", "err")
				part = buildPrinterPart(outFilePath, errFilePath)
				(*g.partsByKey)[partKey] = part
				updated = true
				scannedPrinterPartsKeys = append(scannedPrinterPartsKeys, partKey)
			}
			if part != nil && qualifier == "closed" {
				part.closed = true
				updated = true
				// fmt.Printf("pp: %s/%s marked closed\n", g.name, partKey)
			}
		}
		return err
	})

	allPartsClosed := true
	for k, v := range *g.partsByKey {
		if !collectionz.Contains(&scannedPrinterPartsKeys, k) {
			// part not scanned => must be removed
			delete(*g.partsByKey, k)
			continue
		}
		allPartsClosed = allPartsClosed && v.closed
	}
	g.closed = allPartsClosed

	return updated, agg.Return()
}

type sessionGroup struct {
	path     string
	name     string
	priority int

	notifierParts      *printerGroup
	printersByPrioName *map[string]*printerGroup
	pointer            string
	started, ended     bool
	deadline           *time.Time
	// endMsg             string
}

func (g *sessionGroup) scanFiles() (updated bool, err error) {
	agg := errorz.NewAgg()

	// fmt.Printf("will scan sessionGroup FS: %s ...\n", g.path)
	zcreenFs := os.DirFS(g.path)
	err = fs.WalkDir(zcreenFs, ".", func(path string, d fs.DirEntry, err error) error {
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
				return err
			}
			name := submatches[2]
			key := forgePrioNameKey(priority, name)
			if _, ok := (*g.printersByPrioName)[key]; !ok {
				// If new printer group add it
				group := buildPrinterGroup(path, name, priority)
				(*g.printersByPrioName)[key] = group
			}

			// Do not walk session dir
			return fs.SkipDir
		} else {
			// fmt.Printf("path: %s do not match printerDirFilepathMatcher\n", path)
		}
		return err
	})
	agg.Add(err)

	if g.notifierParts != nil {
		ok, err := g.notifierParts.scanFiles()
		updated = updated || ok
		agg.Add(err)
	}

	for _, v := range *g.printersByPrioName {
		ok, err := v.scanFiles()
		agg.Add(err)
		updated = updated || ok
	}

	return updated, agg.Return()
}

type zcreenGroup struct {
	path               string
	notifierParts      *printerGroup
	sessionsByPrioName *map[string]*sessionGroup
	pointer            string
}

func (g *zcreenGroup) outputer() *zcreenGroupOutputer {
	cursors := make(map[*os.File]int64)
	outputedSessions := make(map[*sessionGroup]bool)
	outputedParts := make(map[*printerPart]bool)
	return &zcreenGroupOutputer{
		zg:               g,
		cursors:          &cursors,
		outputedSessions: &outputedSessions,
		outputedParts:    &outputedParts,
	}
}

func (g *zcreenGroup) scanFiles() (updated bool, err error) {
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
			var ok bool
			var group *sessionGroup

			key := forgePrioNameKey(priority, name)
			if group, ok = (*g.sessionsByPrioName)[key]; !ok {
				group = buildSessionGroup(path, name, priority)
			}
			group.name = name
			group.priority = priority
			(*g.sessionsByPrioName)[key] = group

			// Do not walk session dir
			return fs.SkipDir
		}
		return err
	})

	if g.notifierParts != nil {
		ok, err := g.notifierParts.scanFiles()
		updated = updated || ok
		agg.Add(err)
	}

	sessionUniqness := make(map[string]bool)
	for _, v := range *g.sessionsByPrioName {
		if sessionUniqness[v.name] {
			agg.Add(fmt.Errorf("session: %s is not uniq", v.name))
		}
		ok, err := v.scanFiles()
		updated = updated || ok
		agg.Add(err)
		sessionUniqness[v.name] = true
	}

	return updated, agg.Return()
}

func buildPrinterPart(outFilepath, errFilepath string) *printerPart {
	pp := &printerPart{
		outPath: outFilepath,
		errPath: errFilepath,
		closed:  false,
	}
	if filez.ExistsOrPanic(outFilepath) {
		pp.outFile = filez.OpenReadOnlyOrPanic(outFilepath)
	}
	if filez.ExistsOrPanic(errFilepath) {
		pp.errFile = filez.OpenReadOnlyOrPanic(errFilepath)
	}

	return pp
}

func buildPrinterGroup(printerDir, printerName string, printerPriority int) *printerGroup {
	partsByKey := make(map[string]*printerPart)
	return &printerGroup{
		path:       printerDir,
		name:       printerName,
		priority:   printerPriority,
		partsByKey: &partsByKey,
		pointer:    "",
		closed:     false,
	}
}

func buildSessionGroup(sessionDir, sessionName string, sessionPriority int) *sessionGroup {
	printersByPrioName := make(map[string]*printerGroup)
	return &sessionGroup{
		path:               sessionDir,
		name:               sessionName,
		priority:           sessionPriority,
		notifierParts:      buildPrinterGroup(filepath.Join(sessionDir, notifiersDir), "__notifier", 0),
		printersByPrioName: &printersByPrioName,
		pointer:            "",
		started:            false,
		ended:              false,
		deadline:           nil,
	}
}

func buildZcreenGroup(zcreenDir string) *zcreenGroup {
	sessionsByPrioName := make(map[string]*sessionGroup)
	return &zcreenGroup{
		path:               zcreenDir,
		notifierParts:      buildPrinterGroup(filepath.Join(zcreenDir, notifiersDir), "__notifier", 0),
		sessionsByPrioName: &sessionsByPrioName,
		pointer:            "",
	}
}

// Output all parts of a printer group keeping read context in cursors & outputedParts map.
// If waitForClosed => wait for each printer to be closed before continuing outputing.
func outputsPrinterGroup(cursors *map[*os.File]int64, outputedParts *map[*printerPart]bool, outs printz.Outputs, pg *printerGroup, buf []byte, waitForClosed bool) (err error) {
	printedSomeStuff := false
	allPartsClosed := true
	// fmt.Printf("outputing printer group: %s ...\n", pg.name)
	partsKeys := collectionz.Keys(*pg.partsByKey)
	sort.Strings(partsKeys)

	// First Seek closed printers and promote them top priority
	var closedPpks, openedPpks []string
	for _, ppk := range partsKeys {
		pp := (*pg.partsByKey)[ppk]
		if pp.closed {
			closedPpks = append(closedPpks, ppk)
		} else {
			openedPpks = append(openedPpks, ppk)
		}
	}
	orderedPpks := append(closedPpks, openedPpks...)

	for _, ppk := range orderedPpks {
		pp := (*pg.partsByKey)[ppk]
		if (*outputedParts)[pp] {
			// Do not output already outputed printer parts.
			continue
		} else {
			// fmt.Printf("printer part: %v not already outputed\n", pp)
		}

		// fmt.Printf("outputing printer out part: %s at %d/%d ...\n", pp.outPath, (*cursors)[pp.outFile], (*cursors)[pp.errFile])
		i, j, err := pp.outputAt(outs, (*cursors)[pp.outFile], (*cursors)[pp.errFile], buf)
		if err != nil {
			return fmt.Errorf("error outputAt printer part: %w", err)
		}

		if i > 0 || j > 0 {
			ii := int64(i)
			jj := int64(j)
			(*cursors)[pp.outFile] += ii
			(*cursors)[pp.errFile] += jj
			printedSomeStuff = true
			pg.outputedOut += ii
			pg.outputedErr += jj
			// fmt.Printf("advanced cursor to out: %d err: %d\n", (*cursors)[pp.outFile], (*cursors)[pp.errFile])
		}

		allPartsClosed = allPartsClosed && pp.closed
		if pp.closed {
			// Flag printer part as outputed
			(*outputedParts)[pp] = true
			// fmt.Printf("marked printer outputed\n")
		} else if waitForClosed {
			// fmt.Printf("waiting not closed pg: %s; %v\n", pg.name, pp)
			// current part not closed => exit loop
			break
		}
	}

	if printedSomeStuff {
		err = outs.Flush()
		if err != nil {
			return fmt.Errorf("error flushing outputs: %w", err)
		}
	}
	pg.closed = allPartsClosed

	return
}

type zcreenGroupOutputer struct {
	zg               *zcreenGroup
	cursors          *map[*os.File]int64
	outputedSessions *map[*sessionGroup]bool
	outputedParts    *map[*printerPart]bool
}

// Output all parts of a printer keeping read context at global outputer level.
func (o *zcreenGroupOutputer) outputsGlobalPrinter(outs printz.Outputs, pg *printerGroup, buf []byte, waitForClosed bool) error {
	return outputsPrinterGroup(o.cursors, o.outputedParts, outs, pg, buf, waitForClosed)
}

func (o *zcreenGroupOutputer) update() (err error) {
	_, err = o.zg.scanFiles()
	return
}

func (o *zcreenGroupOutputer) sessions() (names []string) {
	sessionGroupsKeys := collectionz.Keys(*o.zg.sessionsByPrioName)
	sort.Strings(sessionGroupsKeys)
	for _, vk := range sessionGroupsKeys {
		v := (*o.zg.sessionsByPrioName)[vk]
		names = append(names, v.name)
	}
	return
}

func (o *zcreenGroupOutputer) outputs(outs printz.Outputs) (err error) {
	// TODO: implem session timeout
	buf := make([]byte, 1024)

	// Attempt to output global notifier before outputing session
	err = o.outputsGlobalPrinter(outs, o.zg.notifierParts, buf, false)
	return err
}

func (o *zcreenGroupOutputer) sessionOutputer(sessionName string) *sessionGroupOutputer {
	var sg *sessionGroup
	for _, v := range *o.zg.sessionsByPrioName {
		if v.name == sessionName {
			sg = v
			break
		}
	}
	if sg == nil {
		// panic(fmt.Sprintf("no session: %s to output", sessionName))
		// Init an empty session group with 0 priority (which must be updated later if printers are used)
		sDir := forgeSessionDirPath(o.zg.path, sessionName, 0)
		sg = buildSessionGroup(sDir, sessionName, 0)
	}
	cursors := make(map[*os.File]int64)
	outputedPrinters := make(map[*printerGroup]bool)
	outputedParts := make(map[*printerPart]bool)
	return &sessionGroupOutputer{
		zgo:              o,
		sg:               sg,
		cursors:          &cursors,
		outputedPrinters: &outputedPrinters,
		outputedParts:    &outputedParts,
	}

}

type sessionGroupOutputer struct {
	zgo                *zcreenGroupOutputer
	sg                 *sessionGroup
	blockingPrinterKey string
	cursors            *map[*os.File]int64
	outputedPrinters   *map[*printerGroup]bool
	outputedParts      *map[*printerPart]bool
	fullyOutputed      bool
}

func (o *sessionGroupOutputer) update() (err error) {
	// FIXME: remove Update() to be included into HasNext() ?
	updated, err := o.sg.scanFiles()
	// reset fullyOutputed flag if scanFiles updated something.
	o.fullyOutputed = o.fullyOutputed && !updated
	return err
}

func (o *sessionGroupOutputer) hasNext() bool {
	// return true until fullyOutputed
	// FIXME: SHOULD we return false if Update() is needed ?
	// FIXME: SHOULD we call Update() on HasNext() ?
	return !o.fullyOutputed
}

// Output all parts of a printer keeping read context at session outputer level.
func (o *sessionGroupOutputer) outputsSessionPrinter(outs printz.Outputs, pg *printerGroup, buf []byte, waitForClosed bool) error {
	return outputsPrinterGroup(o.cursors, o.outputedParts, outs, pg, buf, waitForClosed)
}

func (o *sessionGroupOutputer) outputs(outs printz.Outputs) (err error) {
	// TODO: implem session timeout
	buf := make([]byte, 1024)
	// err = o.sg.scanFiles()
	// if err != nil {
	// 	return err
	// }

	// Attempt to output global notifier before outputing session
	err = o.zgo.outputsGlobalPrinter(outs, o.zgo.zg.notifierParts, buf, false)
	if err != nil {
		return fmt.Errorf("error outputing global notifs: %w", err)
	}

	printerGroupsKeys := collectionz.Keys(*o.sg.printersByPrioName)
	sort.Strings(printerGroupsKeys)

	// First Seek closed printers and promote them top priority
	closedPgks := make(map[int][]string)
	openedPgks := make(map[int][]string)
	var orderedPgks []string
	for _, pgk := range printerGroupsKeys {
		pg := (*o.sg.printersByPrioName)[pgk]
		if pgk == o.blockingPrinterKey {
			// SKip blocking printer which is in first place
			continue
		} else if pg.closed {
			closedPgks[pg.priority] = append(closedPgks[pg.priority], pgk)
		} else {
			openedPgks[pg.priority] = append(openedPgks[pg.priority], pgk)
		}
	}

	// 1- Blocking printer is always oredered first
	var currentPriority = 0
	if o.blockingPrinterKey != "" {
		pg := (*o.sg.printersByPrioName)[o.blockingPrinterKey]
		currentPriority = pg.priority
		orderedPgks = append(orderedPgks, o.blockingPrinterKey)
	}

	closedPgksPrios := collectionz.Keys(closedPgks)
	sort.Ints(closedPgksPrios)
	openedPgksPrios := collectionz.Keys(openedPgks)
	sort.Ints(openedPgksPrios)

	// 2- Then closed printers with same priority are placed
	for _, priority := range closedPgksPrios {
		pgks := closedPgks[priority]
		if priority == currentPriority {
			sort.Strings(pgks)
			orderedPgks = append(orderedPgks, pgks...)
		}
	}

	// 3- Then closed printers with higher priority are placed
	// FIXME: this can output printer in disorder.
	for _, priority := range closedPgksPrios {
		pgks := closedPgks[priority]
		if priority < currentPriority {
			sort.Strings(pgks)
			orderedPgks = append(orderedPgks, pgks...)
		}
	}

	// 4- Then opened printers with equal priority are placed
	for _, priority := range openedPgksPrios {
		pgks := openedPgks[priority]
		if priority == currentPriority {
			sort.Strings(pgks)
			orderedPgks = append(orderedPgks, pgks...)
		}
	}

	// 5- Then opened printers with higher priority are placed
	// FIXME: this can output printer in disorder.
	for _, priority := range openedPgksPrios {
		pgks := openedPgks[priority]
		if priority < currentPriority {
			sort.Strings(pgks)
			orderedPgks = append(orderedPgks, pgks...)
		}
	}

	// 6- Then closed printers with lower priority are placed
	for _, priority := range closedPgksPrios {
		pgks := closedPgks[priority]
		if priority > currentPriority {
			sort.Strings(pgks)
			orderedPgks = append(orderedPgks, pgks...)
		}
	}

	// 7- Then opened printers with lower priority are placed
	for _, priority := range openedPgksPrios {
		pgks := openedPgks[priority]
		if priority > currentPriority {
			sort.Strings(pgks)
			orderedPgks = append(orderedPgks, pgks...)
		}
	}

	//fmt.Printf("currentPriority: %d, blockingPrinterKey: %s,closedPgks: %s \n", currentPriority, o.blockingPrinterKey, collectionz.Values(closedPgks))
	//fmt.Printf("outputing session: %s, orderedPgks: %s ...\n", o.sg.name, orderedPgks)
	// fmt.Printf("printerGroups: %v\n", o.sg.printersByPrioName)

	fullyOutputted := true
	for _, pgk := range orderedPgks {
		pg := (*o.sg.printersByPrioName)[pgk]
		// Attempt to output session notifier before outputing a printer.
		err = o.outputsSessionPrinter(outs, o.sg.notifierParts, buf, false)
		if err != nil {
			return fmt.Errorf("error outputing session notifs: %w", err)
		}

		if (*o.outputedPrinters)[pg] {
			continue
		} else {
			// fmt.Printf(">> printer group not already outputed: %v\n", pg)
		}

		// scan printer group files before outputing to find more printer parts or update their closed state.
		_, err = pg.scanFiles()
		if err != nil {
			return err
		}

		// fmt.Printf("<<>> outputting pg: %s (pg.closed: %v)\n", forgePrioNameKey(pg.priority, pg.name), pg.closed)
		// Output from elected printer group
		err = o.outputsSessionPrinter(outs, pg, buf, true)
		if err != nil {
			return fmt.Errorf("error outputing session printer: %w", err)
		}
		// fmt.Printf(">>>> pg.closed: %v\n", pg.closed)

		if pg.closed {
			// Flag printer group as outputed
			(*o.outputedPrinters)[pg] = true
			o.blockingPrinterKey = ""
		} else if pg.outputedOut > 0 || pg.outputedErr > 0 {
			// current part partially outputed and not closed => exit loop
			o.blockingPrinterKey = pgk
			fullyOutputted = false
			// fmt.Printf("outputer: %s not fully outputed. printer with key: %s not closed yet\n", o.sg.name, pgk)
			break
		}
	}

	// Attempt to output session notifier after all printer were closed.
	err = o.outputsSessionPrinter(outs, o.sg.notifierParts, buf, false)
	if err != nil {
		return fmt.Errorf("error outputing session notifs: %w", err)
	}

	if o.sg.ended {
		(*o.zgo.outputedSessions)[o.sg] = true
		// attempt to output global notifier after session was ended.
		err = o.zgo.outputsGlobalPrinter(outs, o.zgo.zg.notifierParts, buf, false)
		if err != nil {
			return fmt.Errorf("error outputing global notifs: %w", err)
		}
	}
	o.fullyOutputed = fullyOutputted

	return nil
}
