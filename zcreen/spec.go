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
	- => To avoid problems of files in progress of reading by a tailer deleted by a concurrent screen, MUST delegate
	  session tmp files consolidation to a tailer which will write lock the file then delete all other files.
	  Next tailers will attempt to read session tmp files, then printer & notifiers files if session was reopened and it should be OK.
    - A session can be ended and re-opened, so multiple session tmp files COULD be consolidated.
  - A concurrent tailer on an opened session will tail printers & notifiers files until session is ended.
  - Since multiple printer parts must be consolidated, tailer NEED to order them.
    - 1- Order parts by creation time.
    - 2- Consolidated ended parts first.
	- 3- Consolidate first not eneded part then stop consolidation.
	- Do we need a timeout for a not closed part ?
	- Closing a printer close only it's corresponding part.
    - Printer parts must be ended

## File Layer Implementation
### Initial naïve implementation (v1)
Only one zcreen can write.
Printers & Sessions states are held in session .ser file.

<ZCREEN_TMP_DIR>
├── <SESSION_NAME>.ser
├── _-_notifier-err
├── _-_notifier-out
├── ___session__<SESSION_NAME>
│   ├── <SESSION_NAME>-err-1764765448370487489.1629834888
│   ├── <SESSION_NAME>-err-1764765448402195207.2822699552
│   ├── <SESSION_NAME>-out-1764765448370487489.1778338562
│   ├── <SESSION_NAME>-out-1764765448402195207.138887125
│   ├── _-_notifier-err
│   ├── _-_notifier-out
│   └── ___printers__
│       ├── <P0>__<PRINTER_A_NAME>-err-1764765448370618463.132613444
│       ├── <P0>__<PRINTER_A_NAME>-err-1764765448460120374.655308445
│       ├── <P0>__<PRINTER_A_NAME>-err-1764765448462155450.1849251374
│       ├── <P1>__<PRINTER_B_NAME>-err-1764765448402296979.1246430089
│       ├── <P1>__<PRINTER_B_NAME>-err-1764765448460199228.4158067380
│       ├── <P1>__<PRINTER_B_NAME>-err-1764765448462273041.2430513684
│       ├── [...]
│       ├── <Pk>__<PRINTER_N_NAME>-out-1764765448622031527.1079825214
│       ├── <Pk>__<PRINTER_N_NAME>-out-1764765448627213256.3498790631
│       └── <Pk>__<PRINTER_N_NAME>-out-1764765448630265033.2664619092
└── sync.lock

### New implementation (v2)
Multiple zcreen can write simultaneously each in their own printer files.
Tailer v2 will consolidate each files in a coherent display.
- Allow different screen to write on same session/printer.
- /!\ Must know when a printer / session is closed for consolidation ordering.
- Closing printers & sessions is an information needed only by tailer for consolidation.
- Writes on a reopened printer or session already tailed will not be displayed. => We MUST forbid session or printer reopening unless cleared.
- Should a zcreen B allow print on a printer closed by a zcreen A ? CAN a printer be safely reopened ?
- Should a zcreen B allow print in a session closed by a screen A ? CAN a session be safely reopened ?



<ZCREEN_TMP_DIR>
├── __notifiers
│   ├── err.<TIMESTAMP_1>.<PID_A>
│   ├── err.[...]
│   ├── err.<PID_N>.TIMESTAMP_P
│   ├── out.<TIMESTAMP_1>.<PID_A>
│   ├── out.[...]
│   └── out.<PID_N>.TIMESTAMP_P
└── __sessions
    └── <P0>__<SESSION_NAME>
        ├── __notifiers
        │   ├── err.<TIMESTAMP_1>.<PID_A>
        │   ├── err.[...]
    	│   ├── err.<TIMESTAMP_P>.<PID_N>
        │   ├── out.<TIMESTAMP_1>.<PID_A>
        │   ├── out.[...]
        │   └── out.<TIMESTAMP_P>.<PID_N>
        ├── __printers
        │   ├── <P0>__<PRINTER_A_NAME>
        │   │   ├── err.<TIMESTAMP_1>.<PID_A>
    	│   │   ├── err.[...]
    	│   │   ├── err.<TIMESTAMP_P>.<PID_N>
		│   │   ├── closed.<TIMESTAMP_1>.<PID_A>
    	│   │   ├── closed.[...]
    	│   │   ├── closed.<TIMESTAMP_P>.<PID_N>
    	│   │   ├── out.<TIMESTAMP_1>.<PID_A>
	    │   │   ├── out.[...]
        │   │   └── out.<TIMESTAMP_P>.<PID_N>
        │   ├── <P1>__<PRINTER_B_NAME>
        │   │   ├── err.<TIMESTAMP_1>.<PID_A>
        │   │   ├── err.[...]
        │   │   ├── err.<TIMESTAMP_P>.<PID_N>
		│   │   ├── closed.<TIMESTAMP_1>.<PID_A>
        │   │   ├── closed.[...]
        │   │   ├── closed.<TIMESTAMP_P>.<PID_N>
        │   │   ├── out.<TIMESTAMP_1>.<PID_A>
        │   │   ├── out.[...]
        │   │   └── out.<TIMESTAMP_P>.<PID_N>
        │   └── <Pk>__<PRINTER_N_NAME>
        │       ├── err.<TIMESTAMP_1>.<PID_A>
        │       ├── err.[...]
        │       ├── err.<TIMESTAMP_P>.<PID_N>
		│       ├── closed.<TIMESTAMP_1>.<PID_A>
        │       ├── closed.[...]
        │       ├── closed.<TIMESTAMP_P>.<PID_N>
        │       ├── out.<TIMESTAMP_1>.<PID_A>
        │       ├── out.[...]
        │       └── out.<TIMESTAMP_P>.<PID_N>
        └── __session.ser


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
	sessionDirPrefix0   = "___session__"
	printersDirPrefix0  = "___printers__"
	outFileQualifier    = "out"
	errFileQualifier    = "err"
	closedFileQualifier = "closed"
	outFileNameSuffix0  = "-out"
	errFileNameSuffix0  = "-err"
	bufLen              = 1024
)

const (
	tmpDirFileMode        = 0760
	notifierPrinterName0  = "_-_notifier"
	continuousFlushPeriod = 1 * time.Millisecond
	screenLockFilename    = "screen.lock"
	lockFilename          = "sync.lock" // FIXME: rename syncLockFilename
	fileLockingTimeout    = 2 * time.Second
	tailerDelayTimeout    = 5 * time.Second
)

const (
	serializedExtension0    = ".ser"
	sessionSerialedFilename = "__session.ser"
	extraTimeout            = 20 * time.Millisecond
	noPrintTimeout          = 10 * time.Millisecond
	extraNoPrintTimeout     = 20 * time.Millisecond
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
	ClearSession0(session string) error

	// Clear each sessions workspaces.
	Clear0() error
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

func buildTmpOutputs0(tmpDir, name string) (printz.Outputs, *os.File, *os.File) {
	timestamp := time.Now().UnixNano()

	tmpOutFile, err := os.CreateTemp(tmpDir, buildTmpFilename(name, outFileNameSuffix0, timestamp))
	if err != nil {
		panic(err)
	}
	tmpErrFile, err := os.CreateTemp(tmpDir, buildTmpFilename(name, errFileNameSuffix0, timestamp))
	if err != nil {
		panic(err)
	}
	tmpOutputs := printz.NewOutputs(tmpOutFile, tmpErrFile)
	return tmpOutputs, tmpOutFile, tmpErrFile
}

func buildTmpPrinter0(tmpDir, name string, priorityOrder int, opened bool) *printer0 {
	priorizedName := fmt.Sprintf("%d__%s", priorityOrder, name)
	tmpOutputs, tmpOut, tmpErr := buildTmpOutputs0(tmpDir, priorizedName)
	prtr := printz.New(tmpOutputs)
	closingPrtr := printz.Closing(prtr)
	p := &printer0{
		ClosingPrinter: closingPrtr,
		name:           name,
		tmpOut:         tmpOut,
		tmpErr:         tmpErr,
		// open:           opened,
		priorityOrder: priorityOrder,
	}
	return p
}

func scanAndUpdateTmpPrinters0(tmpDir string, tmpPrinters *map[string]*printer0) error {
	wildcardPath := filepath.Join(tmpDir, "*")
	printersFiles, err := filepath.Glob(wildcardPath)
	if err != nil {
		return err
	}
	// fmt.Printf("<< scanning session %s printerFile: %s\n", tmpDir, printersFiles)
	filenamePattern, err := regexp.Compile(".*/(\\d+)__(.+)(?:" + outFileNameSuffix0 + ").*")
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
				m := make(map[string]*printer0)
				*tmpPrinters = m
			}
			if _, ok := (*tmpPrinters)[printerName]; !ok {
				// printer file does not exists in tmpPrintersMap
				tmpOutputs, tmpOut, tmpErr := buildTmpOutputs0(tmpDir, printerName)
				prtr := printz.New(tmpOutputs)
				closingPrtr := printz.Closing(prtr)
				p := &printer0{
					ClosingPrinter: closingPrtr,
					name:           printerName,
					tmpOut:         tmpOut,
					tmpErr:         tmpErr,
					// open:           false, // A scanned printer cannot be scanned open.
					priorityOrder: priorityOrder,
				}
				(*tmpPrinters)[printerName] = p
				fmt.Printf("<< scanning session printerFile: %s => name: %s #%d\n", printerFile, printerName, priorityOrder)
			}
		}
	}
	return nil
}

func buildPrinter0(tmpDir, name string, priorityOrder int) *printer0 {
	tmpOutFile, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("%s%s", name, outFileNameSuffix0)))
	if err != nil {
		panic(err)
	}
	tmpErrFile, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("%s%s", name, errFileNameSuffix0)))
	if err != nil {
		panic(err)
	}
	tmpOutputs := printz.NewOutputs(tmpOutFile, tmpErrFile)

	prtr := printz.New(tmpOutputs)
	closingPrtr := printz.Closing(prtr)
	p := &printer0{
		ClosingPrinter: closingPrtr,
		name:           name,
		tmpOut:         tmpOutFile,
		tmpErr:         tmpErrFile,
		// open:           true,
		priorityOrder: priorityOrder,
	}
	return p
}

func buildReadOnlyPrinter0(tmpDir, name string, priorityOrder int) *printer0 {
	outNotifierPath := filepath.Join(tmpDir, fmt.Sprintf("%s%s", name, outFileNameSuffix0))
	errNotifierPath := filepath.Join(tmpDir, fmt.Sprintf("%s%s", name, errFileNameSuffix0))
	// filez.WaitUntilExistsOrPanic(outNotifierPath, tailerDelayTimeout)
	// filez.WaitUntilExistsOrPanic(errNotifierPath, tailerDelayTimeout)

	tmpOutFile, err := filez.Open(outNotifierPath, os.O_RDONLY+os.O_CREATE, filez.DefaultFilePerms)
	if err != nil {
		panic(err)
	}
	tmpErrFile, err := filez.Open(errNotifierPath, os.O_RDONLY+os.O_CREATE, filez.DefaultFilePerms)
	if err != nil {
		panic(err)
	}
	tmpOutputs := printz.NewOutputs(tmpOutFile, tmpErrFile)

	prtr := printz.New(tmpOutputs)
	closingPrtr := printz.Closing(prtr)
	p := &printer0{
		ClosingPrinter: closingPrtr,
		name:           name,
		tmpOut:         tmpOutFile,
		tmpErr:         tmpErrFile,
		// open:           true,
		priorityOrder: priorityOrder,
	}
	return p
}
