package zcreen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/printz"
)

/**
- A screen is a structure which abstract the whole screen.
  - a screen may be splitted into multiple parts which are displayed simultaneously.
  - for now only one part is supported to be displayed at once
- A session is a structure which permit to print data on the screen.
  - Multiple sessions can coexist simultaneously
  - Only one sessions is elected at once to be displayed in a screen part until closed
- A printer belong to a session
  - Multiple // printers can coexist simultaneously
  - Printers are flushed in order in session
- A displayer is the interface used to print data on the screen.
  - stdout & stderr


## Features
- A displayer can be configured with Formatters
- The configuration should be stored at Screen level
- Multiple process "share" the display
  - One "main" process will flush the screen on std outputs
  - // process will open sessions and printers
  - One opened session belong to a process => 2 process cannot share one session
  - All printers of a session may be used in // but by the same process.
  - Session printers are concatenated in order


## Flushing v2
  - Flush a printer => write into tmp file
  - Flush a session => concat closed printers in order + currently opened printer into a session tmp file ()
  - Flush a screen => print sessions in order onto std outputs (keep written bytes count)
  - /!\ CHANGE: flush a session MUST not concat closed printer. Flush a session MUST only flush all session printers.
  - /!\ CHANGE: flush screen MUST only flush all sessions.
  - /!\ CHANGE: flush screen tailer MUST concat session closed printers into session tmp files.


## Tailer v2
  - Tailer is now responsible to concat all session files.
  - Sessions files MUST NOT be consolidated on session or screen flush anymore.
  - Session consolidation must gather all available printers and notifiers files.
  - Multiple screen MUST write in their own printer and notifiers files.
  - Tailer MUST track bytes counts read on each files.
  - Tailer COULD have a session serialized on a file.
  - Session tmp files COULD be consolidated on session end, but it's not necessary.
    - /!\ If a session is consolidated, MUST rm all files which were consolidated.
	- A session can be ended and re-opened, so multiple session tmp files can be consolidated.
    - A concurrent tailer on an opened session will tail printers & notifiers files until session is ended.
	- => To avoid problems of files in progress of reading by a tailer deleted by a concurrent screen, MUST delegate
	  session tmp files consolidation to a tailer which will write lock the file then delete all other files.
	  Next tailers will attempt to read session tmp files, then printer & notifiers files if session was reopened and it should be OK.





## TODO
  - Doc tailing
  - Doc notifying
  - Doc why multiple sessions files ? Do we need to add multiple notifier files ?





## Ideas
- reuse printz.Printer : how to format ?
- screen.InitPrinter(name)
- screen.ConfigPrinter(name, formaters)
- change format.Formatter for signature: Format(string|Stringer...) Formatted
- use templates ?
- add notifications push wich will be displayed between sessions
*/

const (
	sessionDirPrefix  = "___session__"
	printersDirPrefix = "___printers__"
	outFileNameSuffix = "-out"
	errFileNameSuffix = "-err"
	bufLen            = 1024
)

const (
	tmpDirFileMode        = 0760
	notifierPrinterName   = "_-_notifier"
	continuousFlushPeriod = 1 * time.Millisecond
	screenLockFilename    = "screen.lock"
	lockFilename          = "sync.lock" // FIXME: rename syncLockFilename
	fileLockingTimeout    = 2 * time.Second
	tailerDelayTimeout    = 5 * time.Second
)

const (
	serializedExtension = ".ser"
	extraTimeout        = 20 * time.Millisecond
)

type Sink interface {
	Session(name string, priority int) (*session, error)
	NotifyPrinter() printz.Printer
	Close() error

	// Continuously Flush supplied session until it's end or timeout is reached.
	FlushBlocking(session string, timeout time.Duration) error

	// Continuously Flush all sessions until ends or timeout is reached.
	FlushAllBlocking(timeout time.Duration) error

	// Resync sink with tailer
	Resync() error
}

type Session interface {
	Printer(name string, priority int) (printz.Printer, error)
	ClosePrinter(name, message string) error
	NotifyPrinter() printz.Printer

	// Consolidate session outputs
	Flush() error

	// Flush printed messages which where not consolidated before session's end.
	Reclaim() error

	// Start the session
	Start(timeout time.Duration, timeoutCallbacks ...func(Session)) error

	// End the session
	End(message string) error
}

type Tailer interface {
	// Continuously Tail supplied session until it's end or timeout is reached.
	// Do not tail notifications
	TailOnlyBlocking(session string, timeout time.Duration) error

	// Continuously Tail opened sessions in order until supplied session's end or timeout is reached.
	// tail notifications before session"s start and after session's end
	TailBlocking(session string, timeout time.Duration) error

	// Continuously Tail all opened sessions in order until ends or timeout is reached.
	// tail notifications between sessions
	TailAllBlocking(timeout time.Duration) error

	// Continuously Tail supplied opened sessions in order until ends or timeout is reached.
	// tail notifications between sessions
	TailSuppliedBlocking(sessionNames []string, timeout time.Duration) error

	// Tail ended session containing some flushed print not tailed.
	Reclaim(session string) error

	// Tail all ended sessions in order which contains some flushed print not tailed.
	ReclaimAll() error

	// Clear session workspace.
	ClearSession(session string) error

	// Clear each sessions workspaces.
	Clear() error
}

func buildTmpFilename(name, qualifier string, timestamp int64) string {
	return fmt.Sprintf("%s%s-%d.*", name, qualifier, timestamp)
}

func tmpFilenameMatches(dir, name, qualifier string) []string {
	wildcardPath := fmt.Sprintf("%s%s-*", name, qualifier)
	matches, err := filepath.Glob(dir + "/" + wildcardPath)
	if err != nil {
		panic(err)
	}
	return matches
}

func priorizedTmpFilenameMatches(dir, name, qualifier string) []string {
	wildcardPath := fmt.Sprintf("*__%s%s-*", name, qualifier)
	matches, err := filepath.Glob(dir + "/" + wildcardPath)
	if err != nil {
		panic(err)
	}
	return matches
}

func buildTmpOutputs(tmpDir, name string) (printz.Outputs, *os.File, *os.File) {
	timestamp := time.Now().UnixNano()

	tmpOutFile, err := os.CreateTemp(tmpDir, buildTmpFilename(name, outFileNameSuffix, timestamp))
	if err != nil {
		panic(err)
	}
	tmpErrFile, err := os.CreateTemp(tmpDir, buildTmpFilename(name, errFileNameSuffix, timestamp))
	if err != nil {
		panic(err)
	}
	tmpOutputs := printz.NewOutputs(tmpOutFile, tmpErrFile)
	return tmpOutputs, tmpOutFile, tmpErrFile
}

func buildTmpPrinter(tmpDir, name string, priorityOrder int, opened bool) *printer {
	priorizedName := fmt.Sprintf("%d__%s", priorityOrder, name)
	tmpOutputs, tmpOut, tmpErr := buildTmpOutputs(tmpDir, priorizedName)
	prtr := printz.New(tmpOutputs)
	closingPrtr := printz.Closing(prtr)
	p := &printer{
		ClosingPrinter: closingPrtr,
		name:           name,
		tmpOut:         tmpOut,
		tmpErr:         tmpErr,
		open:           opened,
		priorityOrder:  priorityOrder,
	}
	return p
}

func scanAndUpdateTmpPrinters(tmpDir string, tmpPrinters *map[string]*printer) error {
	wildcardPath := filepath.Join(tmpDir, "*")
	printersFiles, err := filepath.Glob(wildcardPath)
	if err != nil {
		return err
	}
	// fmt.Printf("<< scanning session %s printerFile: %s\n", tmpDir, printersFiles)
	filenamePattern, err := regexp.Compile(".*/(\\d+)__(.+)(?:" + outFileNameSuffix + ").*")
	if err != nil {
		panic(err)
	}
	for _, printerFile := range printersFiles {
		if filenamePattern.MatchString(printerFile) {
			matches := filenamePattern.FindStringSubmatch(printerFile)
			priority := matches[1]
			printerName := matches[2]
			priorityOrder, err := strconv.Atoi(priority)
			if err != nil {
				return err
			}
			// fmt.Printf("<< scanning session %s printerFile: %s => name: %s #%d\n", tmpDir, printerFile, printerName, priorityOrder)
			if *tmpPrinters == nil {
				m := make(map[string]*printer)
				*tmpPrinters = m
			}
			if _, ok := (*tmpPrinters)[printerName]; !ok {
				// printer file does not exists in tmpPrintersMap
				tmpOutputs, tmpOut, tmpErr := buildTmpOutputs(tmpDir, printerName)
				prtr := printz.New(tmpOutputs)
				closingPrtr := printz.Closing(prtr)
				p := &printer{
					ClosingPrinter: closingPrtr,
					name:           printerName,
					tmpOut:         tmpOut,
					tmpErr:         tmpErr,
					open:           false, // A scanned printer cannot be scanned open.
					priorityOrder:  priorityOrder,
				}
				(*tmpPrinters)[printerName] = p
				fmt.Printf("<< scanning session printerFile: %s => name: %s #%d\n", printerFile, printerName, priorityOrder)
			}
		}
	}
	return nil
}

func buildPrinter(tmpDir, name string, priorityOrder int) *printer {
	tmpOutFile, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("%s%s", name, outFileNameSuffix)))
	if err != nil {
		panic(err)
	}
	tmpErrFile, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("%s%s", name, errFileNameSuffix)))
	if err != nil {
		panic(err)
	}
	tmpOutputs := printz.NewOutputs(tmpOutFile, tmpErrFile)

	prtr := printz.New(tmpOutputs)
	closingPrtr := printz.Closing(prtr)
	p := &printer{
		ClosingPrinter: closingPrtr,
		name:           name,
		tmpOut:         tmpOutFile,
		tmpErr:         tmpErrFile,
		open:           true,
		priorityOrder:  priorityOrder,
	}
	return p
}

func buildReadOnlyPrinter(tmpDir, name string, priorityOrder int) *printer {
	outNotifierPath := filepath.Join(tmpDir, fmt.Sprintf("%s%s", name, outFileNameSuffix))
	errNotifierPath := filepath.Join(tmpDir, fmt.Sprintf("%s%s", name, errFileNameSuffix))
	// filez.WaitUntilExistsOrPanic(outNotifierPath, tailerDelayTimeout)
	// filez.WaitUntilExistsOrPanic(errNotifierPath, tailerDelayTimeout)

	tmpOutFile, err := filez.Open3(outNotifierPath, os.O_RDONLY+os.O_CREATE, filez.DefaultFilePerms)
	if err != nil {
		panic(err)
	}
	tmpErrFile, err := filez.Open3(errNotifierPath, os.O_RDONLY+os.O_CREATE, filez.DefaultFilePerms)
	if err != nil {
		panic(err)
	}
	tmpOutputs := printz.NewOutputs(tmpOutFile, tmpErrFile)

	prtr := printz.New(tmpOutputs)
	closingPrtr := printz.Closing(prtr)
	p := &printer{
		ClosingPrinter: closingPrtr,
		name:           name,
		tmpOut:         tmpOutFile,
		tmpErr:         tmpErrFile,
		open:           true,
		priorityOrder:  priorityOrder,
	}
	return p
}
