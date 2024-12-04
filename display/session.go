package display

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"mby.fr/utils/collections"
	"mby.fr/utils/filez"
	"mby.fr/utils/printz"
)

const (
	serializedExtension = ".ser"
)

type printer struct {
	printz.Printer

	name           string
	opened, closed bool
	priorityOrder  int

	tmpOut, tmpErr       *os.File
	cursorOut, cursorErr int64
}

func serializeSession(s *session) (err error) {
	filePath := filepath.Join(filepath.Dir(s.TmpPath), s.Name+serializedExtension)
	f, err := os.OpenFile(filePath, os.O_CREATE+os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer func() { f.Close() }()
	enc := gob.NewEncoder(f)
	err = enc.Encode(s)
	return
}

func deserializeSession(path string) (s *session, err error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return
	}
	defer func() { f.Close() }()
	dec := gob.NewDecoder(f)
	s = &session{}
	err = dec.Decode(s)
	return s, err
}

type session struct {
	Name           string
	PriorityOrder  int
	Started, Ended bool
	readOnly       bool

	TmpPath                string
	TmpOutName, TmpErrName string
	tmpOut, tmpErr         *os.File
	cursorOut, cursorErr   int64

	// tmpOutputs         map[string]printz.Outputs
	// openedPrinters     map[string]printz.Printer
	printersByPriority map[int][]*printer
	printers           map[string]*printer

	currentPriority *int
}

func (s *session) Printer(name string, priorityOrder int) printz.Printer {
	if !s.Started {
		panic(fmt.Sprintf("session [%s] not started", s.Name))
	}
	if s.Ended {
		panic(fmt.Sprintf("session [%s] ended", s.Name))
	}
	if prtr, ok := s.printers[name]; ok {
		panic(fmt.Sprintf("printer [%s] already exists", name))
		return prtr
	}
	tmpOutputs, tmpOut, tmpErr := buildTmpOutputs(s.TmpPath, name)
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
	err = os.MkdirAll(s.TmpPath, 0755)
	if err != nil {
		panic(err)
	}

	s.Started = true
	// TODO: manage timeout

	err = serializeSession(s)
	return
}

func (s *session) End() (err error) {
	s.Ended = true
	err = serializeSession(s)
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

	printers, ok := s.printersByPriority[*s.currentPriority]
	if ok && len(printers) > 0 {
		openedPrinters := 0
		buf := make([]byte, bufLen)
		for _, prtr := range printers {
			if prtr.closed {
				continue
			} else {
				openedPrinters++
				prtr.Flush()

				n, err := filez.PartialCopy(prtr.tmpOut, s.tmpOut, buf, prtr.cursorOut, -1)
				if err != nil {
					return err
				}
				prtr.cursorOut += int64(n)

				n, err = filez.PartialCopy(prtr.tmpErr, s.tmpErr, buf, prtr.cursorErr, -1)
				if err != nil {
					return err
				}
				prtr.cursorErr += int64(n)
			}
		}

		if openedPrinters == 0 {
			// all printers are closed => clear current priority
			s.currentPriority = nil
		}
	}

	err := serializeSession(s)
	return err
}
