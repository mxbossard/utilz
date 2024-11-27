package display

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"time"

	"mby.fr/utils/collections"
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
)

type displayer interface {
	// Replaced by printz.Printer ?
	Stdout()
	Stderr()
	Flush() error
	//Quiet(bool)
}

type screen struct {
	tmpPath  string
	sessions map[string]*session
	outputs  printz.Outputs
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

func (s *screen) Session(name string, priorityOrder int32) *session {
	if session, ok := s.sessions[name]; ok {
		return session
	}

	sessionDirpath := filepath.Join(s.tmpPath, printerDirPrefix+name)
	if _, err := os.Stat(sessionDirpath); err == nil {
		panic(fmt.Sprintf("unable to create async screen session: [%s] path already exists", sessionDirpath))
	}

	err := os.MkdirAll(sessionDirpath, 0755)
	if err != nil {
		panic(err)
	}

	suiteTmpOutputs, _, _ := buildTmpOutputs(s.tmpPath, name)
	session := &session{
		name:            name,
		priorityOrder:   priorityOrder,
		tmpPath:         sessionDirpath,
		suiteTmpOutputs: suiteTmpOutputs,
		// tmpOutputs:         make(map[string]printz.Outputs),
		// openedPrinters:     make(map[string]printz.Printer),
		printersByPriority: make(map[int32][]*printer),
		printers:           make(map[string]*printer),
	}
	s.sessions[name] = session

	return session
}

func (*screen) ConfigPrinter(name string) (s *session) {
	// TODO later
	return
}

func (*screen) Flush() (err error) {
	// Flush the display
	// TODO
	return
}

func (*screen) AsyncFlush(timeout time.Duration) (err error) {
	// Launch goroutine wich will continuously flush async display
	// TODO
	return
}

func (*screen) BlockTail(timeout time.Duration) (err error) {
	// Tail async display blocking until end
	// TODO
	return
}

type printer struct {
	printz.Printer

	name           string
	opened, closed bool
	priorityOrder  int32

	tmpOut, tmpErr       *os.File
	cursorOut, cursorErr int64
}

type session struct {
	name           string
	priorityOrder  int32
	started, ended bool

	tmpPath         string
	suiteTmpOutputs printz.Outputs

	// tmpOutputs         map[string]printz.Outputs
	// openedPrinters     map[string]printz.Printer
	printersByPriority map[int32][]*printer
	printers           map[string]*printer

	currentPriority *int32
}

func (s *session) Printer(name string, priorityOrder int32) printz.Printer {
	if prtr, ok := s.printers[name]; ok {
		return prtr
	}
	tmpOutputs, tmpOut, tmpErr := buildTmpOutputs(s.tmpPath, name)
	prtr := printz.New(tmpOutputs)
	// s.tmpOutputs[name] = tmpOutputs
	// s.openedPrinters[name] = prtr

	p := &printer{
		Printer:       prtr,
		name:          name,
		tmpOut:        tmpOut,
		tmpErr:        tmpErr,
		opened:        false,
		closed:        false,
		priorityOrder: priorityOrder,
	}
	s.printers[name] = p
	s.printersByPriority[priorityOrder] = append(s.printersByPriority[priorityOrder], p)

	return prtr
}

func (s *session) Close(name string) error {
	// Close a printer
	if prtr, ok := s.printers[name]; ok {
		prtr.closed = true
		err := prtr.tmpOut.Close()
		if err != nil {
			return err
		}
		err = prtr.tmpErr.Close()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no printer opened with name: [%s]", name)
	}
	return nil
}

func (s *session) Start(timeout time.Duration) (err error) {
	s.started = true
	// TODO: manage timeout
	return
}

func (s *session) End() (err error) {
	s.ended = true
	return
}

func (s *session) Flush() error {
	// concat closed printers + current printer into a session tmp file ()

	if s.currentPriority == nil {
		//  Find next priority
		priorityOrders := collections.Keys(&s.printersByPriority)
		if len(priorityOrders) > 0 {
			slices.Sort(priorityOrders)
			for _, priorityOrder := range priorityOrders {
				printers := s.printersByPriority[priorityOrder]
				for _, printer := range printers {
					if !printer.closed {
						// set current priority of first opened printer
						s.currentPriority = &priorityOrders[0]
					}
				}
			}
		}
	}

	if s.currentPriority == nil {
		// Nothing to flush
		return nil
	}

	printers, _ := s.printersByPriority[*s.currentPriority]
	if len(printers) > 0 {
		openedPrinters := 0
		bufLen := 1024
		buf := make([]byte, bufLen)
		for _, prtr := range printers {
			if prtr.closed {
				continue
			} else {
				openedPrinters++
				prtr.Flush()

				n, err := prtr.tmpOut.ReadAt(buf, prtr.cursorOut)
				if err != nil && err != io.EOF {
					return err
				}

				for n > 0 {
					// Loop while buffer is full
					prtr.cursorOut += int64(n)
					k, err := s.suiteTmpOutputs.Out().Write(buf[0:n])
					if err != nil {
						return err
					}
					if k != n {
						err = fmt.Errorf("bytes count read and written mismatch")
						return err
					}
					n, err = prtr.tmpOut.ReadAt(buf, prtr.cursorOut)
					if err != nil && err != io.EOF {
						return err
					}
				}

				n, err = prtr.tmpErr.ReadAt(buf, prtr.cursorErr)
				if err != nil && err != io.EOF {
					return err
				}

				for n > 0 {
					// Loop while buffer is full
					prtr.cursorErr += int64(n)
					k, err := s.suiteTmpOutputs.Err().Write(buf[0:n])
					if err != nil {
						return err
					}
					if k != n {
						err = fmt.Errorf("bytes count read and written mismatch")
						return err
					}
					n, err = prtr.tmpErr.ReadAt(buf, prtr.cursorErr)
					if err != nil && err != io.EOF {
						return err
					}
				}
			}
		}

		if openedPrinters == 0 {
			// all printers are closed => clear current priority
			s.currentPriority = nil
		}
	}

	return nil
}

func NewScreen(outputs printz.Outputs) *screen {
	return &screen{
		sessions: make(map[string]*session),
		outputs:  printz.NewStandardOutputs(),
	}
}

func NewAsyncScreen(outputs printz.Outputs, tmpPath string) *screen {
	if _, err := os.Stat(tmpPath); err == nil {
		panic(fmt.Sprintf("unable to create async screen: [%s] path already exists", tmpPath))
	}

	err := os.MkdirAll(tmpPath, 0760)
	if err != nil {
		panic(err)
	}

	return &screen{
		sessions: make(map[string]*session),
		outputs:  outputs,
		tmpPath:  tmpPath,
	}
}
