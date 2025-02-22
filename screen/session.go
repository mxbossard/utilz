package screen

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"mby.fr/utils/collections"
	"mby.fr/utils/filez"
	"mby.fr/utils/printz"
)

const (
	serializedExtension = ".ser"
	extraTimeout        = 20 * time.Millisecond
)

type printer struct {
	printz.Printer

	name           string
	opened, closed bool
	flushed        bool
	priorityOrder  int

	tmpOut, tmpErr       *os.File
	cursorOut, cursorErr int64
}

// FIXME: use a different struct for serialization with exported fields.
type session struct {
	mutex *sync.Mutex

	Name            string
	PriorityOrder   int
	Started, Ended  bool
	flushed, tailed bool
	readOnly        bool
	timeouted       *time.Duration

	TmpPath                string
	TmpOutName, TmpErrName string
	tmpOut, tmpErr         *os.File
	cursorOut, cursorErr   int64

	printersByPriority map[int][]*printer
	printers           map[string]*printer
	notifier           *printer

	currentPriority *int
}

func (s *session) Printer(name string, priorityOrder int) printz.Printer {
	if name == "" {
		panic("cannot get printer of empty name")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.timeouted != nil {
		panic(fmt.Sprintf("cannot get printer for session [%s] timeouted after: %s", s.Name, *s.timeouted))
	}

	// FIXME: printer should be getable & writable if session not started yet
	if !s.Started {
		panic(fmt.Sprintf("cannot get printer for session [%s] not started", s.Name))
	}
	if s.Ended {
		panic(fmt.Sprintf("cannot get printer for session [%s] already ended", s.Name))
	}
	if prtr, ok := s.printers[name]; ok {
		return prtr
	}

	p := buildTmpPrinter(s.TmpPath, name, priorityOrder)
	s.printers[name] = p
	s.printersByPriority[priorityOrder] = append(s.printersByPriority[priorityOrder], p)

	return p.Printer
}

func (s *session) closePrinter(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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
	err := serializeSession(s)
	return err
}

func (s *session) ClosePrinter(name string) error {
	return s.closePrinter(name)
}

func (s *session) NotifyPrinter() printz.Printer {
	return s.notifier
}

func (s *session) Start(timeout time.Duration, timeoutCallbacks ...func()) (err error) {
	if s.timeouted != nil {
		return fmt.Errorf("cannot start session [%s] timeouted after: %s", s.Name, *s.timeouted)
	}
	if s.Ended {
		return fmt.Errorf("cannot start session: [%s] already ended", s.Name)
	}
	if s.Started {
		// already started
		return nil
	}
	err = os.MkdirAll(s.TmpPath, 0755)
	if err != nil {
		panic(err)
	}

	screenDirPath := filepath.Dir(s.TmpPath)
	_, tmpOut, tmpErr := buildTmpOutputs(screenDirPath, s.Name)
	s.TmpOutName = tmpOut.Name()
	s.TmpErrName = tmpErr.Name()
	s.tmpOut = tmpOut
	s.tmpErr = tmpErr
	s.notifier = buildPrinter(s.TmpPath, notifierPrinterName, 0)

	s.Started = true
	go func() {
		time.Sleep(timeout + extraTimeout)
		if !s.Ended {
			s.timeouted = &timeout
			err := s.End()
			if err != nil {
				panic(err)
			}
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
		err = s.closePrinter(prtr.name)
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

func (s *session) Clear() (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.Ended {
		return fmt.Errorf("cannot clear session: [%s] not ended", s.Name)
	}
	fmt.Printf("Clearing session dir: [%s] ...\n", s.TmpPath)
	s.Started = false
	s.Ended = false
	s.flushed = false

	if s.tmpOut != nil {
		s.tmpOut.Close()
		s.tmpOut = nil
	}
	if s.tmpErr != nil {
		s.tmpErr.Close()
		s.tmpErr = nil
	}

	filePath := sessionSerializedPath(filepath.Dir(s.TmpPath), s.Name)
	err = os.RemoveAll(filePath)
	if err != nil {
		return err
	}

	if s.TmpOutName != "" {
		err = os.RemoveAll(s.TmpOutName)
		if err != nil {
			return err
		}
	}
	if s.TmpErrName != "" {
		err = os.RemoveAll(s.TmpErrName)
		if err != nil {
			return err
		}
	}

	// Attempt to close & remove notifier temp files
	s.notifier.tmpOut.Close()
	s.notifier.tmpErr.Close()
	os.RemoveAll(s.notifier.tmpOut.Name())
	os.RemoveAll(s.notifier.tmpErr.Name())

	err = os.RemoveAll(s.TmpPath)
	if err != nil {
		return err
	}

	s.cursorOut = 0
	s.cursorErr = 0
	s.printersByPriority = make(map[int][]*printer)
	s.printers = make(map[string]*printer)

	return err
}

// Consolidate session outputs with supplied printer content
func (s *session) flushPrinter(p printer) error {
	// TODO
	return nil
}

// Consolidate session outputs with all printers & notifier available
func (s *session) flushAll(name string) error {
	// TODO
	return nil
}

func (s *session) flush() error {
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
			err := s.flush()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *session) Flush() error {
	if s.Ended {
		//return fmt.Errorf("session: [%s] already ended", s.Name)
		return nil
	}

	pt := logger.PerfTimer("session", s.Name)
	defer pt.End()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.flush()
	if err != nil {
		return nil
	}

	s.flushed = true
	err = serializeSession(s)

	return err
}

func buildSession(name string, priorityOrder int, screenDirPath string) *session {
	sessionDirpath := filepath.Join(screenDirPath, sessionDirPrefix+name)
	if _, err := os.Stat(sessionDirpath); err == nil {
		panic(fmt.Sprintf("unable to create async screen session: [%s] path already exists", sessionDirpath))
	}

	session := &session{
		mutex:              &sync.Mutex{},
		Name:               name,
		PriorityOrder:      priorityOrder,
		readOnly:           false,
		TmpPath:            sessionDirpath,
		printersByPriority: make(map[int][]*printer),
		printers:           make(map[string]*printer),
		notifier:           buildPrinter(screenDirPath, notifierPrinterName+"-"+name, 0),
	}

	return session
}

func updateSession(exists *session, filePath string) error {
	session, err := deserializeSession(filePath)
	if err != nil {
		return err
	}
	session.currentPriority = nil
	session.tmpOut = nil
	session.tmpErr = nil
	exists.Started = session.Started
	exists.Ended = session.Ended

	logger.Debug("updated session", "session", *exists)

	return nil
}

func sessionSerializedPath(dir, name string) string {
	filePath := filepath.Join(dir, name+serializedExtension)
	return filePath
}

func serializeSession(s *session) (err error) {
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
		return nil, fmt.Errorf("error opening ser file (%s): %w", path, err)
	}
	defer func() { f.Close() }()
	dec := gob.NewDecoder(f)
	s = &session{}
	err = dec.Decode(s)
	logger.Debug("deserialized session", "filepath", path)
	return s, err
}
