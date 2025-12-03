package zcreen

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	notifiersDir = "__notifiers"
	sessionsDir  = "__sessions"
	printersDir  = "__printers"
)

func forgeSessionDirPath(zcreenPath, sessionName string, sessionPriority int) string {
	name := fmt.Sprintf("%d__%s", sessionPriority, sessionName)
	return filepath.Join(zcreenPath, sessionsDir, name)
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

func forgeUniqContextSuffix() string {
	pid := os.Getpid
	timestamp := time.Now().UnixNano()
	suffix := fmt.Sprintf(".%d.%21d", pid, timestamp)
	return suffix
}

func forgeNotiferPrinterFilepathes(zcreenPath, sessionName string, sessionPriority int) (string, string) {
	dir := forgeNotifiersDirPath(zcreenPath, sessionName, sessionPriority)
	suffix := forgeUniqContextSuffix()
	outFilepath := filepath.Join(dir, fmt.Sprint("out%s", suffix))
	errFilepath := filepath.Join(dir, fmt.Sprint("err%s", suffix))
	return outFilepath, errFilepath
}

func forgeSessionPrinterFilepathes(zcreenPath, sessionName string, sessionPriority int, printerName string, printerPriority int) (string, string) {
	dir := forgePrintersDirPath(zcreenPath, sessionName, sessionPriority)
	suffix := forgeUniqContextSuffix()
	outFilepath := filepath.Join(dir, fmt.Sprint("%d__%s__out%s", printerPriority, printerName, suffix))
	errFilepath := filepath.Join(dir, fmt.Sprint("%d__%s__err%s", printerPriority, printerName, suffix))
	return outFilepath, errFilepath
}

type printerFiles struct {
	outPath   string
	outCursor int64
	outFile   *os.File
	errPath   string
	errCursor int64
	errFile   *os.File
}

type printerGroup struct {
	printerName string
	priority    int

	printers []printerFiles
	pointer  int
	closed   bool
	// closeMsg string
}

type sessionGroup struct {
	path        string
	sessionName string
	priority    int

	notifiers          []printerFiles
	printersByPriority map[int][]printerGroup
	pointer            int
	started, ended     bool
	deadline           time.Time
	// endMsg             string
}

type zcreenGroup struct {
	path               string
	notifiers          []printerFiles
	sessionsByPriority map[int][]sessionGroup
}

func (z *zcreenGroup) scanFiles() error {
	// TODO: scan z.path and add missing sessionGroups, printerGroups and printerFiles.
	// TODO: update printers and sessions states

	return nil
}

func buildZcreenGroup(zcreenPath string) *zcreenGroup {
	return &zcreenGroup{path: zcreenPath}
}
