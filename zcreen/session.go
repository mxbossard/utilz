package zcreen

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/mxbossard/utilz/collectionz"
	"github.com/mxbossard/utilz/errorz"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/printz"
	"github.com/mxbossard/utilz/utilz"
)

// FIXME: use a different struct for serialization with exported fields.
type sessionSerializer struct {
	Name           string
	PriorityOrder  int
	Started, Ended bool
	EndMessage     string
	Timeouted      *time.Duration
	Deadline       *time.Time
	Start, End     *time.Time
	TmpPath        string
}

type session struct {
	mutex *sync.Mutex

	Name             string
	PriorityOrder    int
	Started, Ended   bool
	EndMessage       string
	printed          bool
	flushed, tailed  bool
	cleared          bool
	readOnly         bool
	Timeouted        *time.Duration
	timeoutCallbacks []func(Session)

	TmpPath               string
	serializationFilepath string
	ScreenDirPath         string

	// FIXME: should replace following by a printer but cannot do it simply because files must be used by tailer !
	// tmpOutName, tmpErrName string
	// tmpOut, tmpErr         *os.File
	// cursorOut, cursorErr   int64

	printersByPriority map[int][]*fsPrinter
	printers           map[string]*fsPrinter
	notifier           *fsPrinter

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
	// For concurrent reason on who should open the session, and who will get the first printer, do not forbit to get a printer on a not started session.
	// if !s.Started {
	// 	return nil, fmt.Errorf("session: [%s] not started yet", s.Name)
	// }

	if s.Ended {
		return nil, fmt.Errorf("cannot get printer: [%s / #%d] for session: [%s] already ended with message: %s", name, priorityOrder, s.Name, s.EndMessage)
	}

	if s.Timeouted != nil {
		return nil, fmt.Errorf("cannot get printer: [%s / #%d] for session: [%s] timeouted after: %s", name, priorityOrder, s.Name, *s.Timeouted)
	}

	if prtr, ok := s.printers[name]; ok {
		// fmt.Printf("found printer %s in cache\n", name)
		return prtr, nil
	}

	// printerDirPath := printersDirPath(s.TmpPath)
	// p := buildTmpPrinter(printerDirPath, name, priorityOrder, true)
	p := buildSessionPrinter(s.ScreenDirPath, s.Name, s.PriorityOrder, name, priorityOrder)
	s.printers[name] = p
	s.printersByPriority[priorityOrder] = append(s.printersByPriority[priorityOrder], p)
	// fmt.Printf("built new printer %s\n", name)

	return p, nil
}

func (s *session) closePrinter(name, message string) error {
	// Closing the printer do not close the underlying files
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Mark a printer closed, but do not close backing tmp files.
	if prtr, ok := s.printers[name]; ok {
		if prtr.IsClosed() {
			// Printer already closed
			return nil
		}
		err := prtr.Close(message)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no printer opened with name: [%s] closing message: %s", name, message)
	}
	return nil
}

func (s *session) ClosePrinter(name, message string) error {
	return s.closePrinter(name, message)
}

func (s *session) NotifyPrinter() printz.Printer {
	return s.notifier
}

func (s *session) Start(timeout time.Duration, timeoutCallbacks ...func(Session)) (err error) {
	if s.Timeouted != nil {
		return fmt.Errorf("cannot start session [%s] timeouted after: %s", s.Name, *s.Timeouted)
	}
	if s.Ended {
		// Allow a session to be restarted
		//return fmt.Errorf("cannot start session: [%s] already ended with message: %s", s.Name, s.EndMessage)
	}
	if s.Started {
		// already started
		return nil
	}

	s.timeoutCallbacks = timeoutCallbacks
	s.Started = true
	s.Ended = false
	s.EndMessage = ""
	s.printed = false
	s.flushed = false
	s.tailed = false
	s.cleared = false
	s.Timeouted = nil
	go func() {
		time.Sleep(timeout + extraTimeout)
		if !s.Ended {
			for _, tcb := range s.timeoutCallbacks {
				tcb(s)
			}
			err := s.End(fmt.Sprintf("reach session timeout after %s", timeout))
			s.Timeouted = &timeout // Set timeout state after ending session
			if err != nil {
				logger.Error(err.Error())
				//panic(err)
			}
		}
	}()

	err = serializeSession(s)
	return
}

// Close a session : cannot write anymore in it but the session is not ended (could be reopen)
func (s *session) close(message string) (err error) {
	// close all opened printers
	for _, prtr := range s.printers {
		err = prtr.Close(message)
		if err != nil {
			return err
		}
	}

	if s.notifier != nil {
		err = s.notifier.Close(message)
		if err != nil {
			return err
		}
	}

	// FIXME: we shoud wait for session tail to flush sessions ?
	// err = s.Flush()
	// if err != nil {
	// 	return err
	// }

	// FIXME: we shoud wait for session tail to consolidate notifications ?
	// Consolidate notifications "after suite"
	// err = s.consolidateNotifier()
	// if err != nil {
	// 	return err
	// }

	// if s.timeouted == nil {
	// 	err = serializeSession(s)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (s *session) End(message string) (err error) {
	// if !s.Started {
	// 	return fmt.Errorf("cannot end not yet started session: %s with message: %s", s.Name, message)
	// }
	if s.Ended {
		return
	}

	s.close(message)

	// Must End after close since close have a different comportment if session is ended.
	s.Ended = true
	s.EndMessage = message

	if s.Timeouted == nil {
		err = serializeSession(s)
		if err != nil {
			return err
		}
	}
	logger.Debug("session ended", "session", s.Name, "message", message)

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

	// if s.tmpOut != nil {
	// 	s.tmpOut.Close()
	// 	s.tmpOut = nil
	// }
	// if s.tmpErr != nil {
	// 	s.tmpErr.Close()
	// 	s.tmpErr = nil
	// }

	// if s.tmpOutName != "" {
	// 	err = os.RemoveAll(s.tmpOutName)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// if s.tmpErrName != "" {
	// 	err = os.RemoveAll(s.tmpErrName)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// Attempt to close & remove notifier temp files
	if s.notifier != nil && !s.notifier.IsClosed() {
		err = s.notifier.Close(fmt.Sprintf("cleared session: %s", s.Name))
		if err != nil {
			return err
		}
	}

	// err = os.RemoveAll(sessionSerializedPath(s.TmpPath))
	// if err != nil {
	// 	return err
	// }

	err = os.RemoveAll(s.TmpPath)
	if err != nil {
		return err
	}

	// s.cursorOut = 0
	// s.cursorErr = 0

	//s.Name = ""
	//s.PriorityOrder = 0
	//s.TmpPath = ""
	// s.ScreenDirPath = ""
	s.Ended = false
	s.Started = false
	s.EndMessage = ""
	s.printed = false
	s.flushed = false
	s.tailed = false
	s.readOnly = false
	s.Timeouted = nil
	s.timeoutCallbacks = nil
	s.serializationFilepath = ""
	s.currentPriority = nil
	s.printersByPriority = make(map[int][]*fsPrinter)
	s.printers = make(map[string]*fsPrinter)
	s.notifier = buildNotifierPrinter(s.ScreenDirPath, s.Name, s.PriorityOrder)
	s.timeoutCallbacks = nil
	s.cleared = true

	printersDirPath := printersDirPath(s.TmpPath)
	err = os.MkdirAll(printersDirPath, filez.DefaultDirPerms)
	if err != nil {
		err = fmt.Errorf("unable to create async screen session: [%s] dir: %w", printersDirPath, err)
		return err
	}

	// err = serializeSession(s)

	return err
}

func (s *session) Clear() (err error) {
	return s.clear()
}

// Consolidate session outputs with supplied printer content
// concat printer into a session tmp file
// func (s *session) consolidatePrinter(prtr *printer) error {
// 	err := prtr.Flush()
// 	if err != nil {
// 		return err
// 	}

// 	buf := make([]byte, bufLen)
// 	if !prtr.open && prtr.consolidated {
// 		// fmt.Printf("printer closed: [%s]\n", prtr.name)
// 		return nil
// 	} else {
// 		// fmt.Printf("flushing printer: [%s] ; cursor: [%d] ; flushed: [%v] ; closed: [%v]\n", prtr.name, prtr.cursorOut, prtr.flushed, prtr.closed)
// 		err := prtr.Flush()
// 		if err != nil {
// 			return err
// 		}

// 		n, err := filez.CopyChunk(prtr.tmpOut, s.tmpOut, buf, prtr.cursorOut, -1)
// 		if err != nil {
// 			return err
// 		}
// 		prtr.cursorOut += int64(n)

// 		n, err = filez.CopyChunk(prtr.tmpErr, s.tmpErr, buf, prtr.cursorErr, -1)
// 		if err != nil {
// 			return err
// 		}
// 		prtr.cursorErr += int64(n)
// 	}
// 	prtr.consolidated = true
// 	// fmt.Printf("flushed printer: [%s] ; cursor: [%d] ; flushed: [%v] ; closed: [%v]\n", prtr.name, prtr.cursorOut, prtr.flushed, prtr.closed)

// 	if !prtr.open {
// 		err = prtr.Close(fmt.Sprintf("printer: %s closed with message: %s", prtr.name, prtr.closeMessage))
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// Consolidate session outputs with supplied printer content
// func (s *session) consolidateNotifier() error {
// 	if s.notifier != nil && s.notifier.IsClosed() {
// 		// FIXME: should check if notifier files are closed, not the printer
// 		return nil
// 	}
// 	return s.consolidatePrinter(s.notifier)
// }

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
					// if !printer.consolidated && !printer.LastPrint().IsZero() {
					if !printer.LastPrint().IsZero() {
						// set current priority of first opened printer
						s.printed = true // mark session print began
						s.currentPriority = &priorityOrder
						break out
					}
					nothingPrintedYet = nothingPrintedYet && !printer.IsClosed() && printer.LastPrint().IsZero()
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
// func (s *session) consolidateAll() error {
// 	s.nextPriority()

// 	if s.currentPriority == nil && !s.printed {
// 		// While no printer was consolidated yet, Consolidate notifications "before suite"
// 		err := s.consolidateNotifier()
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	for s.currentPriority != nil {
// 		printers, ok := s.printersByPriority[*s.currentPriority]
// 		if !ok || len(printers) == 0 {
// 			break
// 		}

// 		openedPrinters := 0
// 		// buf := make([]byte, bufLen)
// 		for _, prtr := range printers {
// 			err := s.consolidatePrinter(prtr)
// 			if err != nil {
// 				return err
// 			}

// 			if prtr.open {
// 				openedPrinters++
// 			}
// 		}

// 		// fmt.Printf("opened printers: [%d]\n", openedPrinters)
// 		if openedPrinters == 0 {
// 			// all printers are closed => clear current priority
// 			s.currentPriority = nil
// 			s.nextPriority()
// 		} else {
// 			break
// 		}
// 	}

// 	return nil
// }

func (s *session) Flush() error {
	if s.Ended {
		//return fmt.Errorf("session: [%s] already ended", s.Name)
		return nil
	}

	pt := logger.PerfTimer("session", s.Name)
	defer pt.End()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	agg := errorz.NewAggregated()
	// Flush notifier
	if s.notifier != nil {
		err := s.notifier.Flush()
		agg.Add(err)
	}

	// Flush all printers
	for _, prtr := range s.printers {
		err := prtr.Flush()
		agg.Add(err)
	}

	// err := s.consolidateAll()
	// if err != nil {
	// 	return nil
	// }

	//s.flushed = true
	//err = serializeSession(s)

	return agg.Return()
}

func (s *session) Reclaim() error {
	// TODO
	panic("not implemented yet")
}

func (s *session) refresh() error {
	//TODO
	if s.printers == nil {
		s.printers = make(map[string]*fsPrinter)
	}
	if s.printersByPriority == nil {
		s.printersByPriority = make(map[int][]*fsPrinter)
	}
	for _, printer := range s.printers {
		s.printersByPriority[printer.priorityOrder] = append(s.printersByPriority[printer.priorityOrder], printer)
	}
	for _, printers := range s.printersByPriority {
		slices.SortStableFunc(printers, func(a, b *fsPrinter) int {
			return strings.Compare(a.name, b.name)
		})
	}
	return nil
}

// Refresh session from FS layer
func (s *session) refresh0() error {
	printerDirPath := printersDirPath(s.TmpPath)
	wildcardPath := filepath.Join(printerDirPath, "*")
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
			printerPriority, err := strconv.Atoi(priority)
			printerName := matches[2]
			if err != nil {
				return err
			}
			// fmt.Printf("<< scanning session %s printerFile: %s => name: %s #%d\n", printerDirPath, printerFile, printerName, printerPriority)
			if s.printers == nil {
				s.printers = make(map[string]*fsPrinter)
			}
			if _, ok := s.printers[printerName]; !ok {
				// printer file does not exists in tmpPrintersMap
				// tmpOutputs, tmpOut, tmpErr := buildTmpOutputs(printerDirPath, printerName)
				// prtr := printz.New(tmpOutputs)
				// prtr := buildTmpPrinter(printerDirPath, printerName, printerPriority, false)
				prtr := buildSessionPrinter(s.ScreenDirPath, s.Name, s.PriorityOrder, printerName, printerPriority)
				// closingPrtr := printz.Closing(prtr)
				// p := &printer{
				// 	ClosingPrinter: closingPrtr,
				// 	name:           printerName,
				// 	tmpOut:         tmpOut,
				// 	tmpErr:         tmpErr,
				// 	open:           false, // A scanned printer cannot be scanned open.
				// 	priorityOrder:  priorityOrder,
				// }
				s.printers[printerName] = prtr
				// fmt.Printf("<< scanning session printerFile: %s => name: %s #%d\n", printerFile, printerName, printerPriority)
			}
		}
	}

	// Clear and rebuild priority map
	s.printersByPriority = make(map[int][]*fsPrinter)
	for _, printer := range s.printers {
		s.printersByPriority[printer.priorityOrder] = append(s.printersByPriority[printer.priorityOrder], printer)
	}
	for _, printers := range s.printersByPriority {
		slices.SortStableFunc(printers, func(a, b *fsPrinter) int {
			return strings.Compare(a.name, b.name)
		})
	}
	return nil
}

func buildSession(name string, priorityOrder int, screenDirPath string) (s *session, err error) {
	sessionDirPath := forgeSessionDirPath(screenDirPath, name, priorityOrder)
	sessionSerPath := sessionSerializedPath(sessionDirPath)
	if _, err := os.Stat(sessionSerPath); err == nil {
		//return nil, fmt.Errorf("unable to create async screen session: [%s] path already exists", sessionDirpath)
		// session path already exists
		s, err = deserializeSession(sessionSerPath)
		if err != nil {
			return nil, err
		}
		s.PriorityOrder = priorityOrder
		// Session must be restarted correctly
		s.Started = false
		//session.Ended = false
		s.readOnly = false
		s.printersByPriority = make(map[int][]*fsPrinter)
		s.printers = make(map[string]*fsPrinter)
	} else {
		s = &session{
			mutex:              &sync.Mutex{},
			Name:               name,
			PriorityOrder:      priorityOrder,
			TmpPath:            sessionDirPath,
			readOnly:           false,
			printersByPriority: make(map[int][]*fsPrinter),
			printers:           make(map[string]*fsPrinter),
			ScreenDirPath:      screenDirPath,
		}
	}

	// Init session pointers
	printersDirPath := printersDirPath(sessionDirPath)
	err = os.MkdirAll(printersDirPath, filez.DefaultDirPerms)
	if err != nil {
		err = fmt.Errorf("unable to create async screen session: [%s] dir: %w", printersDirPath, err)
		return nil, err
	}

	// _, tmpOut, tmpErr := buildTmpOutputs(sessionDirPath, s.Name)
	// s.tmpOutName = tmpOut.Name()
	// s.tmpErrName = tmpErr.Name()
	// s.tmpOut = tmpOut
	// s.tmpErr = tmpErr
	// s.notifier = buildPrinter(sessionDirPath, notifierPrinterName, 0)
	s.notifier = buildNotifierPrinter(s.ScreenDirPath, s.Name, s.PriorityOrder)

	return s, nil
}

func updateSession(exists *session, filePath string) error {
	session, err := deserializeSession(filePath)
	if err != nil {
		return fmt.Errorf("unable to update session: %w", err)
	}
	session.currentPriority = nil
	// session.tmpOut = nil
	// session.tmpErr = nil
	exists.Started = session.Started
	exists.Ended = session.Ended
	exists.EndMessage = session.EndMessage

	logger.Debug("updated session", "session", *exists)

	return nil
}

func sessionDirPath0(screenDirPath, sessionName string) string {
	return filepath.Join(screenDirPath, sessionDirPrefix0+sessionName)
}

func printersDirPath(sessionDirPath string) string {
	return filepath.Join(sessionDirPath, printersDir)
}

func sessionSerializedPath(sessionDir string) string {
	filePath := filepath.Join(sessionDir, sessionSerialedFilename)
	return filePath
}

func serializeSession(s *session) (err error) {
	if s.cleared {
		// Do not serialize cleared sessions
		return nil
	}

	filePath := sessionSerializedPath(s.TmpPath)
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
	s.printersByPriority = make(map[int][]*fsPrinter)
	s.printers = make(map[string]*fsPrinter)
	err = dec.Decode(s)
	if s.TmpPath == "" {
		panic("empty session TmpPath")
	}
	// s.notifier = buildPrinter(s.TmpPath, notifierPrinterName, 0)
	// s.notifier = buildNotifierPrinter(s.ScreenDirPath, s.Name, s.PriorityOrder, true)

	logger.Debug("deserialized session", "name", s.Name, "filepath", path)
	return s, err
}

func clearSessionFiles(zcreenPath, sessionName string) error {
	sessionsDir := forgeSessionsDirPath(zcreenPath)
	sessionDirsWildcard := filepath.Join(sessionsDir, "*__"+sessionName)
	matchingDirs, err := filepath.Glob(sessionDirsWildcard)
	// fmt.Printf("clearSessionFiles: %s; matchingDirs: %s\n", sessionDirsWildcard, matchingDirs)
	if err != nil {
		return err
	}
	if len(matchingDirs) == 0 {
		return fmt.Errorf("no session with name: %s found", sessionName)
	} else if len(matchingDirs) > 1 {
		return fmt.Errorf("more than one session with name: %s found", sessionName)
	}

	// sessionDirPath := sessionDirPath(zcreenPath, sessionName)
	sessionDirPath := matchingDirs[0]
	if _, err := os.Stat(sessionDirPath); err == nil {
		err := os.RemoveAll(sessionDirPath)
		if err != nil {
			return fmt.Errorf("cannot remove session dir: %w", err)
		}
		logger.Debug("Tailer: cleared session dir", "dir", sessionDirPath)
	}
	serPath := sessionSerializedPath(sessionDirPath)
	if _, err := os.Stat(serPath); err == nil {
		err = os.RemoveAll(serPath)
		if err != nil {
			return fmt.Errorf("cannot remove session ser file: %w", err)
		}
		logger.Debug("Tailer: cleared session ser file", "serPath", serPath)
	}
	return nil
}
