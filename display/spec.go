package display

import (
	"fmt"
	"os"

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
*/

const (
	printerDirPrefix  = "printers_"
	outFileNameSuffix = "-out.*"
	errFileNameSuffix = "-err.*"
	bufLen            = 1024
)

type displayer interface {
	// Replaced by printz.Printer ?
	Stdout()
	Stderr()
	Flush() error
	//Quiet(bool)
}

func buildTmpOutputs(tmpDir, name string) (printz.Outputs, *os.File, *os.File) {
	tmpOutFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("%s"+outFileNameSuffix, name))
	if err != nil {
		panic(err)
	}
	tmpErrFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("%s"+errFileNameSuffix, name))
	if err != nil {
		panic(err)
	}
	tmpOutputs := printz.NewOutputs(tmpOutFile, tmpErrFile)
	return tmpOutputs, tmpOutFile, tmpErrFile
}
