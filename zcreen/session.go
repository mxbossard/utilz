package zcreen

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/mxbossard/utilz/collectionz"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/printz"
	"github.com/mxbossard/utilz/utilz"
)

type printer struct {
	printz.ClosingPrinter

	name string
	// open bool
	open          bool
	closeMessage  string
	consolidated  bool
	priorityOrder int

	tmpOut, tmpErr       *os.File
	cursorOut, cursorErr int64
}

func (p *printer) Close(message string) error {
	err := p.ClosingPrinter.Close(message)
	if err != nil {
		return err
	}

	if p.tmpOut != nil {
		err := p.tmpOut.Close()
		if err != nil {
			return err
		}
		err = os.RemoveAll(p.tmpOut.Name())
		if err != nil {
			return err
		}
	}

	if p.tmpErr != nil {
		err := p.tmpErr.Close()
		if err != nil {
			return err
		}
		err = os.RemoveAll(p.tmpErr.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

// FIXME: use a different struct for serialization with exported fields.
type session struct {
	mutex *sync.Mutex

	Name             string
	PriorityOrder    int
	Started, Ended   bool
	endMessage       string
	printed          bool
	flushed, tailed  bool
	cleared          bool
	readOnly         bool
	timeouted        *time.Duration
	timeoutCallbacks []func(Session)

	TmpPath               string
	serializationFilepath string

	// FIXME: should replace following by a printer but cannot do it simply because files must be used by tailer !
	tmpOutName, tmpErrName string
	tmpOut, tmpErr         *os.File
	cursorOut, cursorErr   int64
	oldTmpSessionsScanned  bool

	printersByPriority map[int][]*printer
	printers           map[string]*printer
	notifier           *printer

	currentPriority *int
}

func (s session) String() string {
	return fmt.Sprintf("{Session> name: %s; Started: %v; Ended: %v; Cleared: %v}", s.Name, s.Started, s.Ended, s.cleared)
}

func (s *session) Printer(name string, priorityOrder int) (printz.Printer, error) {
	if name == "" {
		return nil, fmt.Errorf("cannot get printer of empty name")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// FIXME: should printer be getable & writable if session not started yet ?
	if !s.Started {
		return nil, fmt.Errorf("session: [%s] not started yet", s.Name)
	}

	if s.Ended {
		return nil, fmt.Errorf("session: [%s] already ended with message: %s", s.Name, s.endMessage)
	}

	if s.timeouted != nil {
		return nil, fmt.Errorf("cannot get printer for session [%s] timeouted after: %s", s.Name, *s.timeouted)
	}

	if prtr, ok := s.printers[name]; ok {
		return prtr, nil
	}

	printerDirPath := printersDirPath(s.TmpPath)
	p := buildTmpPrinter(printerDirPath, name, priorityOrder)
	s.printers[name] = p
	s.printersByPriority[priorityOrder] = append(s.printersByPriority[priorityOrder], p)

	return p.ClosingPrinter, nil
}

func (s *session) closePrinter(name, message string) error {
	// Closing the printer do not close the underlying files
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Mark a printer closed, but do not close backing tmp files.
	if prtr, ok := s.printers[name]; ok {
		if !prtr.open {
			// Printer already closed
			return nil
		}
		// set consolidated to false to enforce a final consolidation
		prtr.consolidated = false
		prtr.open = false
		prtr.closeMessage = message

	} else {
		return fmt.Errorf("no printer opened with name: [%s]", name)
	}
	// err := serializeSession(s)
	// return err
	return nil
}

func (s *session) ClosePrinter(name, message string) error {
	return s.closePrinter(name, message)
}

func (s *session) NotifyPrinter() printz.Printer {
	return s.notifier
}

func (s *session) Start(timeout time.Duration, timeoutCallbacks ...func(Session)) (err error) {
	if s.timeouted != nil {
		return fmt.Errorf("cannot start session [%s] timeouted after: %s", s.Name, *s.timeouted)
	}
	if s.Ended {
		return fmt.Errorf("cannot start session: [%s] already ended with message: %s", s.Name, s.endMessage)
	}
	if s.Started {
		// already started
		return nil
	}

	//sessionDirPath := filepath.Dir(s.TmpPath)
	sessionDirPath := s.TmpPath
	printersDirPath := printersDirPath(sessionDirPath)
	err = os.MkdirAll(printersDirPath, filez.DefaultDirPerms)
	if err != nil {
		err = fmt.Errorf("unable to create async screen session: [%s] dir: %w", printersDirPath, err)
		return
	}

	_, tmpOut, tmpErr := buildTmpOutputs(sessionDirPath, s.Name)
	s.tmpOutName = tmpOut.Name()
	s.tmpErrName = tmpErr.Name()
	s.tmpOut = tmpOut
	s.tmpErr = tmpErr
	s.notifier = buildPrinter(sessionDirPath, notifierPrinterName, 0)
	//buildPrinter(sessionDirpath, notifierPrinterName+"-"+name, 0),

	s.timeoutCallbacks = timeoutCallbacks
	s.Started = true
	go func() {
		time.Sleep(timeout + extraTimeout)
		if !s.Ended {
			s.timeouted = &timeout
			for _, tcb := range s.timeoutCallbacks {
				tcb(s)
			}
			err := s.End(fmt.Sprintf("reach session timeout after %s", timeout))
			if err != nil {
				logger.Error(err.Error())
				//panic(err)
			}
		}
	}()

	err = serializeSession(s)
	return
}

func (s *session) End(message string) (err error) {
	if s.Ended {
		return
	}
	s.endMessage = message

	// close all opened printers
	for _, prtr := range s.printers {
		err = s.closePrinter(prtr.name, message)
		if err != nil {
			return err
		}
	}

	err = s.Flush()
	if err != nil {
		return err
	}

	// FIXME: we shoud wait for session end to consolidate notifications ?
	// Consolidate notifications "after suite"
	err = s.consolidateNotifier()
	if err != nil {
		return err
	}

	s.Ended = true
	if s.timeouted == nil {
		err = serializeSession(s)
		if err != nil {
			return err
		}
	}
	logger.Debug("session ended", "session", s.Name)

	return
}

func (s *session) clear() (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.Ended {
		return fmt.Errorf("cannot clear session: [%s] not ended", s.Name)
	}
	//fmt.Printf("Clearing session dir: [%s] ...\n", s.TmpPath)

	//s.Started = false
	//s.Ended = false

	//s.flushed = false

	if s.tmpOut != nil {
		s.tmpOut.Close()
		s.tmpOut = nil
	}
	if s.tmpErr != nil {
		s.tmpErr.Close()
		s.tmpErr = nil
	}

	if s.tmpOutName != "" {
		err = os.RemoveAll(s.tmpOutName)
		if err != nil {
			return err
		}
	}
	if s.tmpErrName != "" {
		err = os.RemoveAll(s.tmpErrName)
		if err != nil {
			return err
		}
	}

	// Attempt to close & remove notifier temp files
	if s.notifier != nil {
		err = s.notifier.Close(fmt.Sprintf("cleared session: %s", s.Name))
		if err != nil {
			return err
		}
	}

	err = os.RemoveAll(s.TmpPath)
	if err != nil {
		return err
	}

	s.cursorOut = 0
	s.cursorErr = 0
	s.printersByPriority = make(map[int][]*printer)
	s.printers = make(map[string]*printer)
	s.cleared = true

	err = serializeSession(s)

	return err
}

// Consolidate session outputs with supplied printer content
// concat printer into a session tmp file
func (s *session) consolidatePrinter(prtr *printer) error {
	err := prtr.Flush()
	if err != nil {
		return err
	}

	buf := make([]byte, bufLen)
	if !prtr.open && prtr.consolidated {
		// fmt.Printf("printer closed: [%s]\n", prtr.name)
		return nil
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
	prtr.consolidated = true
	// fmt.Printf("flushed printer: [%s] ; cursor: [%d] ; flushed: [%v] ; closed: [%v]\n", prtr.name, prtr.cursorOut, prtr.flushed, prtr.closed)

	if !prtr.open {
		err = prtr.Close(fmt.Sprintf("printer: %s closed with message: %s", prtr.name, prtr.closeMessage))
		if err != nil {
			return err
		}
	}

	return nil
}

// Consolidate session outputs with supplied printer content
func (s *session) consolidateNotifier() error {
	if s.notifier.IsClosed() {
		// FIXME: should check if notifier files are closed, not the printer
		return nil
	}
	return s.consolidatePrinter(s.notifier)
}

func (s *session) nextPriority() {
	if s.currentPriority == nil {
		//  Find next priority
		priorityOrders := collectionz.Keys(s.printersByPriority)
		if len(priorityOrders) > 0 {
			slices.Sort(priorityOrders)
		out:
			for _, priorityOrder := range priorityOrders {
				printers := s.printersByPriority[priorityOrder]
				nothingPrintedYet := true
				for _, printer := range printers {
					// fmt.Printf("selecting printer: [%s] ? prio: [%d]\n", printer.name, priorityOrder)
					if !printer.consolidated && !printer.LastPrint().IsZero() {
						// set current priority of first opened printer
						s.printed = true // mark session print began
						s.currentPriority = &priorityOrder
						break out
					}
					nothingPrintedYet = nothingPrintedYet && printer.open && printer.LastPrint().IsZero()
				}
				// if nothing printed yet for current priority => do not select priority
				if nothingPrintedYet {
					break out
				}
			}
		}
	}
}

// Consolidate session outputs with all printers & notifier available
// concat all closed printers + current printer into a session tmp file
func (s *session) consolidateAll() error {
	s.nextPriority()

	if s.currentPriority == nil && !s.printed {
		// While no printer was consolidated yet, Consolidate notifications "before suite"
		err := s.consolidateNotifier()
		if err != nil {
			return err
		}
	}

	for s.currentPriority != nil {
		printers, ok := s.printersByPriority[*s.currentPriority]
		if !ok || len(printers) == 0 {
			break
		}

		openedPrinters := 0
		// buf := make([]byte, bufLen)
		for _, prtr := range printers {
			err := s.consolidatePrinter(prtr)
			if err != nil {
				return err
			}

			if prtr.open {
				openedPrinters++
			}
		}

		// fmt.Printf("opened printers: [%d]\n", openedPrinters)
		if openedPrinters == 0 {
			// all printers are closed => clear current priority
			s.currentPriority = nil
			s.nextPriority()
		} else {
			break
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

	err := s.consolidateAll()
	if err != nil {
		return nil
	}

	//s.flushed = true
	//err = serializeSession(s)

	return err
}

func (s *session) Reclaim() error {
	// TODO
	panic("not implemented yet")
}

func buildSession(name string, priorityOrder int, screenDirPath string) (*session, error) {
	sessionSerPath := sessionSerializedPath(screenDirPath, name)
	if _, err := os.Stat(sessionSerPath); err == nil {
		//return nil, fmt.Errorf("unable to create async screen session: [%s] path already exists", sessionDirpath)
		// session path already exists
		session, err := deserializeSession(sessionSerPath)
		if err != nil {
			return nil, err
		}
		session.PriorityOrder = priorityOrder
		// Session must be restarted correctly
		session.Started = false
		session.Ended = false
		session.readOnly = false
		session.printersByPriority = make(map[int][]*printer)
		session.printers = make(map[string]*printer)
		return session, nil
	}

	sessionDirPath := sessionDirPath(screenDirPath, name)
	session := &session{
		mutex:              &sync.Mutex{},
		Name:               name,
		PriorityOrder:      priorityOrder,
		TmpPath:            sessionDirPath,
		readOnly:           false,
		printersByPriority: make(map[int][]*printer),
		printers:           make(map[string]*printer),
	}

	return session, nil
}

func updateSession(exists *session, filePath string) error {
	session, err := deserializeSession(filePath)
	if err != nil {
		return fmt.Errorf("unable to update session: %w", err)
	}
	session.currentPriority = nil
	session.tmpOut = nil
	session.tmpErr = nil
	exists.Started = session.Started
	exists.Ended = session.Ended

	logger.Debug("updated session", "session", *exists)

	return nil
}

func sessionDirPath(screenDirPath, sessionName string) string {
	return filepath.Join(screenDirPath, sessionDirPrefix+sessionName)
}

func printersDirPath(sessionDirPath string) string {
	return filepath.Join(sessionDirPath, printersDirPrefix)
}

func sessionSerializedPath(dir, name string) string {
	filePath := filepath.Join(dir, name+serializedExtension)
	return filePath
}

func serializeSession(s *session) (err error) {
	if s.cleared {
		// Do not serialize cleared sessions
		return nil
	}

	filePath := sessionSerializedPath(filepath.Dir(s.TmpPath), s.Name)
	fl := flock.New(filePath)
	err = utilz.FileLock(fl, fileLockingTimeout)
	if err != nil {
		return
	}
	defer utilz.FileUnlock(fl)

	f, err := os.OpenFile(filePath, os.O_CREATE+os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer func() { f.Close() }()
	enc := gob.NewEncoder(f)
	err = enc.Encode(s)
	logger.Debug("serialized session", "name", s.Name, "filepath", filePath)
	return
}

func deserializeSession(path string) (s *session, err error) {
	fl := flock.New(path)
	err = utilz.FileLock(fl, fileLockingTimeout)
	if err != nil {
		return
	}
	defer utilz.FileUnlock(fl)

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("error opening ser file (%s): %w", path, err)
	}
	defer func() { f.Close() }()
	dec := gob.NewDecoder(f)
	s = &session{mutex: &sync.Mutex{}}
	err = dec.Decode(s)
	logger.Debug("deserialized session", "name", s.Name, "filepath", path)
	return s, err
}

func clearSessionFiles(zcreenPath, sessionName string) error {
	sessionDirPath := sessionDirPath(zcreenPath, sessionName)
	if _, err := os.Stat(sessionDirPath); err == nil {
		err := os.RemoveAll(sessionDirPath)
		if err != nil {
			return fmt.Errorf("cannot remove session dir: %w", err)
		}
		logger.Debug("Tailer: cleared session dir", "dir", sessionDirPath)
	}
	serPath := sessionSerializedPath(zcreenPath, sessionName)
	if _, err := os.Stat(serPath); err == nil {
		err = os.RemoveAll(serPath)
		if err != nil {
			return fmt.Errorf("cannot remove session ser file: %w", err)
		}
		logger.Debug("Tailer: cleared session ser file", "serPath", serPath)
	}
	return nil
}
