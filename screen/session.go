package screen

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

	name                    string
	opened, closed, flushed bool
	priorityOrder           int

	tmpOut, tmpErr       *os.File
	cursorOut, cursorErr int64
}

func serializedPath0(s *session) string {
	filePath := filepath.Join(filepath.Dir(s.TmpPath), s.Name+serializedExtension)
	return filePath
}

func sessionSerializedPath(dir, name string) string {
	filePath := filepath.Join(dir, name+serializedExtension)
	return filePath
}

func serializeSession(s *session) (err error) {
	//filePath := serializedPath0(s)
	filePath := sessionSerializedPath(filepath.Dir(s.TmpPath), s.Name)
	f, err := os.OpenFile(filePath, os.O_CREATE+os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer func() { f.Close() }()
	enc := gob.NewEncoder(f)
	err = enc.Encode(s)
	logger.Debug("serialized session", "filepath", filePath)
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
	logger.Debug("deserialized session", "filepath", path)
	return s, err
}

// FIXME: use a different struct for serialization with exported fields.
type session struct {
	Name                    string
	PriorityOrder           int
	Started, Ended, flushed bool
	readOnly                bool

	TmpPath                string
	TmpOutName, TmpErrName string
	tmpOut, tmpErr         *os.File
	cursorOut, cursorErr   int64

	printersByPriority map[int][]*printer
	printers           map[string]*printer

	currentPriority *int
}

func (s *session) Printer(name string, priorityOrder int) printz.Printer {
	if !s.Started {
		panic(fmt.Sprintf("session [%s] not started", s.Name))
	}
	if s.Ended {
		panic(fmt.Sprintf("session [%s] already ended", s.Name))
	}
	if _, ok := s.printers[name]; ok {
		panic(fmt.Sprintf("printer [%s] already exists", name))
		//return prtr
	}

	p := buildTmpPrinter(s.TmpPath, name, priorityOrder)
	s.printers[name] = p
	s.printersByPriority[priorityOrder] = append(s.printersByPriority[priorityOrder], p)

	return p.Printer
}

func (s *session) ClosePrinter(name string) error {
	if s.Ended {
		return fmt.Errorf("session: [%s] already ended", s.Name)
	}
	// Mark a printer closed, but do not close backing tmp files.
	if prtr, ok := s.printers[name]; ok {
		if prtr.closed {
			// Printer already closed
			return nil
		}
		// set flushed to false to enforce a final flush
		prtr.flushed = false
		prtr.closed = true

	} else {
		return fmt.Errorf("no printer opened with name: [%s]", name)
	}
	//fmt.Printf("Closed printer: [%s]\n", name)
	err := serializeSession(s)
	return err
}

func (s *session) Start(timeout time.Duration) (err error) {
	if s.Started {
		return fmt.Errorf("session: [%s] already started", s.Name)
	}
	if s.Ended {
		return fmt.Errorf("session: [%s] already ended", s.Name)
	}
	err = os.MkdirAll(s.TmpPath, 0755)
	if err != nil {
		panic(err)
	}

	s.Started = true
	go func() {
		time.Sleep(timeout)
		err := s.End()
		if err != nil {
			panic(err)
		}
	}()

	err = serializeSession(s)
	return
}

func (s *session) End() (err error) {
	if s.Ended {
		return
	}

	// close all opened printers
	for _, prtr := range s.printers {
		err = s.ClosePrinter(prtr.name)
		if err != nil {
			return err
		}
	}

	err = s.Flush()
	if err != nil {
		return err
	}

	s.Ended = true
	err = serializeSession(s)
	logger.Debug("session ended", "session", s.Name)

	return
}

func (s *session) Flush() error {
	pt := logger.PerfTimer("session", s.Name)
	defer pt.End()

	if s.Ended {
		//return fmt.Errorf("session: [%s] already ended", s.Name)
		return nil
	}
	// concat closed printers + current printer into a session tmp file ()
	if s.currentPriority == nil {
		//  Find next priority
		priorityOrders := collections.Keys(s.printersByPriority)
		if len(priorityOrders) > 0 {
			slices.Sort(priorityOrders)
		out:
			for _, priorityOrder := range priorityOrders {
				printers := s.printersByPriority[priorityOrder]
				for _, printer := range printers {
					// fmt.Printf("selecting printer: [%s] ? prio: [%d]\n", printer.name, priorityOrder)
					if !printer.flushed {
						// set current priority of first opened printer
						s.currentPriority = &priorityOrder
						break out
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
			if prtr.closed && prtr.flushed {
				// fmt.Printf("printer closed: [%s]\n", prtr.name)
				continue
			} else {
				// fmt.Printf("flushing printer: [%s] ; cursor: [%d] ; flushed: [%v] ; closed: [%v]\n", prtr.name, prtr.cursorOut, prtr.flushed, prtr.closed)
				err := prtr.Flush()
				if err != nil {
					return err
				}

				n, err := filez.CopyChunk(prtr.tmpOut, s.tmpOut, buf, prtr.cursorOut, -1)
				if err != nil {
					return err
				}
				prtr.cursorOut += int64(n)

				n, err = filez.CopyChunk(prtr.tmpErr, s.tmpErr, buf, prtr.cursorErr, -1)
				if err != nil {
					return err
				}
				prtr.cursorErr += int64(n)
			}
			prtr.flushed = true
			// fmt.Printf("flushed printer: [%s] ; cursor: [%d] ; flushed: [%v] ; closed: [%v]\n", prtr.name, prtr.cursorOut, prtr.flushed, prtr.closed)
			if prtr.closed {
				//fmt.Printf("closing printer: [%s] in Flush()\n", prtr.name)
				// Close files
				err := prtr.tmpOut.Close()
				if err != nil {
					return err
				}
				err = prtr.tmpErr.Close()
				if err != nil {
					return err
				}
			} else {
				openedPrinters++
			}
		}

		// fmt.Printf("opened printers: [%d]\n", openedPrinters)
		if openedPrinters == 0 {
			// all printers are closed => clear current priority
			s.currentPriority = nil
			// FIXME: do not use a recursive call.
			err := s.Flush()
			if err != nil {
				return err
			}
		}
	}

	s.flushed = true
	err := serializeSession(s)

	return err
}

func buildSession(name string, priorityOrder int, screenDirPath string) *session {
	sessionDirpath := filepath.Join(screenDirPath, sessionDirPrefix+name)
	if _, err := os.Stat(sessionDirpath); err == nil {
		panic(fmt.Sprintf("unable to create async screen session: [%s] path already exists", sessionDirpath))
	}

	_, tmpOut, tmpErr := buildTmpOutputs(screenDirPath, name)
	session := &session{
		Name:               name,
		PriorityOrder:      priorityOrder,
		readOnly:           false,
		TmpPath:            sessionDirpath,
		TmpOutName:         tmpOut.Name(),
		TmpErrName:         tmpErr.Name(),
		tmpOut:             tmpOut,
		tmpErr:             tmpErr,
		printersByPriority: make(map[int][]*printer),
		printers:           make(map[string]*printer),
	}

	return session
}
