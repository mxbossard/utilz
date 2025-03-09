package screen

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mby.fr/utils/printz"
)

/**
- A screen is a structure which abstract the whole screen.
  - a screen may be splitted into multiple parts which are displayed simultaneously.
  - for now only one part is supported to be displayed at once
- A session is a structure which permit to print data on the screen.
  - Multiple sessions can coexist simultaneously
  - Only one sessions is elected at once to be displayed until closed
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


## Flushing
  - Flush a printer => write into tmp file
  - Flush a session => concat closed printers + current printer into a session tmp file ()
  - Flush a screen => print sessions in order onto std outputs (keep written bytes count)



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
	outFileNameSuffix = "-out.*"
	errFileNameSuffix = "-err.*"
	bufLen            = 1024
)

type Sink interface {
	Session(name string, priority int) *session
	ClearSession(name string) error
	NotifyPrinter() printz.Printer
	Close() error

	// Continuously Flush supplied session until it's end or timeout is reached.
	FlushBlocking(session string, timeout time.Duration) error

	// Continuously Flush all sessions until ends or timeout is reached.
	FlushAllBlocking(timeout time.Duration) error

	// Clear each sessions workspaces.
	//Clear() error
}

type Session interface {
	Printer(name string, priority int) printz.Printer
	ClosePrinter(name string) error
	NotifyPrinter() printz.Printer

	// Consolidate session outputs
	Flush() error

	// Flush printed messages which where not consolidated before session's end.
	Reclaim() error

	// Start the session
	Start(timeout time.Duration, timeoutCallbacks ...func(Session)) error

	// End the session
	End() error

	//Clear() error
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

	// Tail ended session containing some flushed print not tailed.
	Reclaim(session string) error

	// Tail all ended sessions in order which contains some flushed print not tailed.
	ReclaimAll() error

	// Clear session workspace.
	ClearSession(session string) error

	// Clear each sessions workspaces.
	Clear() error
}

func buildTmpOutputs(tmpDir, name string) (printz.Outputs, *os.File, *os.File) {
	tmpOutFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("%s%s", name, outFileNameSuffix))
	if err != nil {
		panic(err)
	}
	tmpErrFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("%s%s", name, errFileNameSuffix))
	if err != nil {
		panic(err)
	}
	tmpOutputs := printz.NewOutputs(tmpOutFile, tmpErrFile)
	return tmpOutputs, tmpOutFile, tmpErrFile
}

func buildTmpPrinter(tmpDir, name string, priorityOrder int) *printer {
	tmpOutputs, tmpOut, tmpErr := buildTmpOutputs(tmpDir, name)
	prtr := printz.New(tmpOutputs)
	p := &printer{
		Printer:       prtr,
		name:          name,
		tmpOut:        tmpOut,
		tmpErr:        tmpErr,
		opened:        false,
		closed:        false,
		priorityOrder: priorityOrder,
	}
	return p
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
	p := &printer{
		Printer:       prtr,
		name:          name,
		tmpOut:        tmpOutFile,
		tmpErr:        tmpErrFile,
		opened:        false,
		closed:        false,
		priorityOrder: priorityOrder,
	}
	return p
}
