package zcreen

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/mxbossard/utilz/collectionz"
	"github.com/mxbossard/utilz/errorz"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/printz"
	"github.com/mxbossard/utilz/utilz"
	"github.com/mxbossard/utilz/zlog"
)

var (
	buf    = make([]byte, bufLen)
	logger = zlog.New()
)

type screen struct {
	sync.Mutex
	screenLock *flock.Flock
	fileLock   *flock.Flock
	tmpPath    string
	sessions   map[string]*session
	notifier   *printer
	closed     bool
}

func (s *screen) Session(name string, priorityOrder int) (*session, error) {
	if name == "" {
		panic("cannot get session of empty name")
	}
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return nil, fmt.Errorf("cannot create a session in a closed zcreen")
	}
	if session, ok := s.sessions[name]; ok {
		return session, nil
	}

	session, err := buildSession(name, priorityOrder, s.tmpPath)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("Built sink session: [%s]\n", name)
	s.sessions[name] = session
	return session, nil
}

func (s *screen) NotifyPrinter() printz.Printer {
	return s.notifier.ClosingPrinter
}

func (s *screen) FlushBlocking(sessionName string, timeout time.Duration) (err error) {
	pt := logger.PerfTimer("sessionName", sessionName)
	defer pt.End()

	s.Lock()
	defer s.Unlock()
	startTime := time.Now()
	if session, ok := s.sessions[sessionName]; ok {
		for !session.Ended {
			if time.Since(startTime) > timeout {
				err := errorz.Timeoutf(timeout, "FlushBlocking() for session: [%s]", sessionName)
				return err
			}
			err = session.Flush()
			if err != nil {
				return err
			}
			time.Sleep(continuousFlushPeriod)
		}
	}
	return
}

func (s *screen) FlushAllBlocking(timeout time.Duration) (err error) {
	pt := logger.PerfTimer()
	defer pt.End()

	s.Lock()
	defer s.Unlock()
	startTime := time.Now()
	notEndedCount := -1
	for notEndedCount != 0 {
		notEndedCount = 0
		for _, ses := range s.sessions {
			if time.Since(startTime) > timeout {
				allSessions := collectionz.Values(s.sessions)
				notEndedNames := collectionz.Map(&allSessions, func(s *session) string { return s.Name })
				err := errorz.Timeoutf(timeout, "FlushAllBlocking() some sessions: [%s]", notEndedNames)
				return err
			}
			if !ses.Ended {
				notEndedCount++
				err = ses.Flush()
				if err != nil {
					return err
				}
			}
		}
	}
	return
}

func (s *screen) Close() (err error) {
	s.Lock()
	defer s.Unlock()
	defer s.screenLock.Unlock()
	err = s.notifier.tmpOut.Close()
	if err != nil {
		return err
	}
	err = s.notifier.tmpErr.Close()

	var agg errorz.Aggregated
	for _, s := range s.sessions {
		// Closing screen SHOULD not end sessions but close it : In case of zcreen failure, a restart must take back the session not ended.
		agg.Add(s.close("screen closed"))
	}
	s.closed = true
	return agg.Return()
}

func (s *screen) Resync() error {
	s.Lock()
	defer s.Unlock()
	err := utilz.FileLock(s.fileLock, fileLockingTimeout)
	if err != nil {
		return err
	}
	defer utilz.FileUnlock(s.fileLock)

	scannedSessions, err := scanSerializedSessions(s.tmpPath)
	if err != nil {
		return fmt.Errorf("error scanning sessions: %w", err)
	}

	sessionsToRemove := []string{}
FirstLoop:
	for _, session := range s.sessions {
		for _, scanned := range scannedSessions {
			if session.Name == scanned.Name {
				//fmt.Printf("\n<<>> RESYNC: refreshing %s => %s\n", scanned, session)
				// refresh session ?
				// session.cleared = scanned.cleared
				session.Started = scanned.Started
				session.Ended = scanned.Ended
				logger.Debug("Resync: kept session", "session", session.Name)
				continue FirstLoop
			}
		}
		// session was not scanned : remove it
		sessionsToRemove = append(sessionsToRemove, session.Name)

	}

	for _, sessionName := range sessionsToRemove {
		//fmt.Printf("\n<<>> RESYNC: removing session: %s\n", sessionName)
		clearSessionsMap(&s.sessions, sessionName)
		logger.Debug("Resync: removed session", "session", sessionName)
	}
	return nil
}

/*
func (s *screen) ClearSession(name string) error {
	s.Lock()
	defer s.Unlock()
	//fmt.Printf("Clearing sink session: [%s] (count before: %d)...\n", name, len(s.sessions))
	err := clearSessionsMap(&s.sessions, name)
	//fmt.Printf("Cleared sink session: [%s] (count after: %d)...\n", name, len(s.sessions))
	return err
}

func (s *screen) Clear() (err error) {
	s.Lock()
	defer s.Unlock()

	for _, session := range s.sessions {
		err := session.Clear()
		if err != nil {
			return err
		}
	}

	// Attempt to close & remove notifier temp files
	s.notifier.tmpOut.Close()
	s.notifier.tmpErr.Close()
	os.RemoveAll(s.notifier.tmpOut.Name())
	os.RemoveAll(s.notifier.tmpErr.Name())

	err = os.RemoveAll(s.tmpPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(s.tmpPath, tmpDirFileMode)
	if err != nil {
		return err
	}

	s.sessions = make(map[string]*session)
	s.notifier = buildPrinter(s.tmpPath, notifierPrinterName, 0)
	return
}
*/

func (*screen) ConfigPrinter(name string) (s *session) {
	// TODO later
	return
}

func scanSerializedSessions(zcreenPath string) (sessions []*session, err error) {
	wildcard := sessionSerializedPath(zcreenPath, "*")
	sers, err := filepath.Glob(wildcard)
	if err != nil {
		err = fmt.Errorf("unable to scan ser files: %w", err)
		return
	}

	for _, filePath := range sers {
		var scanned *session
		scanned, err = deserializeSession(filePath)
		if err != nil {
			err = fmt.Errorf("unable to process ser file: %w", err)
			return
		}
		scanned.serializationFilepath = filePath
		// printerDirPath := printersDirPath(scanned.TmpPath)
		// err = scanAndUpdateTmpPrinters(printerDirPath, &scanned.printers)
		err = scanned.refresh()
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, scanned)
	}
	return
}

func hasNextSessionTmp(ses *session) (ok bool, nextTmpOut, nextTmpErr string) {
	matches, _ := filepath.Glob(ses.TmpPath + "/" + ses.Name + outFileNameSuffix + "*")
	if len(matches) > 0 {
		sort.Strings(matches)
		isNext := false
		for _, match := range matches {
			if ses.tmpOutName == "" || isNext {
				nextTmpOut = match
				ok = true
				break
			} else if match == ses.tmpOutName {
				// found current one
				isNext = true
			}
		}
	}

	matches, _ = filepath.Glob(ses.TmpPath + "/" + ses.Name + errFileNameSuffix + "*")
	if len(matches) > 0 {
		sort.Strings(matches)
		isNext := false
		for _, match := range matches {
			if ses.tmpErrName == "" || isNext {
				nextTmpErr = match
				ok = true
				break
			} else if match == ses.tmpErrName {
				// found current one
				isNext = true
			}
		}
	}

	return
}

// Update session tmp path and file in case of tmp file change (if sink restart for example)
// Always take next tmp files (ordered by timestamped name)
func (s *screenTailer) updateNextSessionTmp(ses *session) (updated bool, err error) {
	if !ses.cleared {
		ok, nextOut, nextErr := hasNextSessionTmp(ses)
		if ok {
			if nextOut != "" {
				if ses.tmpOut != nil {
					err = ses.tmpOut.Close()
					if err != nil {
						return
					}
				}
				ses.tmpOutName = nextOut
				ses.tmpOut, err = os.OpenFile(nextOut, os.O_RDONLY, 0)
				if err != nil {
					err = fmt.Errorf("error opening session out tmp file (%s): %w", ses.tmpOutName, err)
					return
				}

				// Clear cursor
				ses.cursorOut = 0
				ses.tailed = false
				ses.flushed = false
				ses.cleared = false
				ses.printed = false
				updated = true
			}
			if nextErr != "" {
				if ses.tmpErr != nil {
					err = ses.tmpErr.Close()
					if err != nil {
						return
					}
				}
				ses.tmpErrName = nextErr
				ses.tmpErr, err = os.OpenFile(nextErr, os.O_RDONLY, 0)
				if err != nil {
					err = fmt.Errorf("error opening session err tmp file (%s): %w", ses.tmpErrName, err)
					return
				}
				// Clear cursor
				ses.cursorErr = 0
				ses.tailed = false
				ses.flushed = false
				ses.cleared = false
				ses.printed = false
				updated = true
			}
		}
	}

	return
}

func (s *screenTailer) scanSessions() (err error) {
	pt := logger.PerfTimer("tmpPath", s.tmpPath)
	defer pt.End("sessionsCount", len(s.sessions))

	s.Lock()
	defer s.Unlock()

	err = utilz.FileLock(s.fileLock, fileLockingTimeout)
	if err != nil {
		return
	}
	defer utilz.FileUnlock(s.fileLock)

	scannedSession, err := scanSerializedSessions(s.tmpPath)
	if err != nil {
		return fmt.Errorf("unable to scan ser files: %w", err)
	}

	for _, scanned := range scannedSession {
		var exists *session
		var ok bool
		// clear session
		scanned.currentPriority = nil
		scanned.tmpOut = nil
		scanned.tmpErr = nil
		if exists, ok = s.sessions[scanned.Name]; !ok {
			// init session
			s.sessions[scanned.Name] = scanned
			s.sessionsByPriority[scanned.PriorityOrder] = append(s.sessionsByPriority[scanned.PriorityOrder], scanned)
			exists = scanned
		} else {
			// update session
			exists.Started = scanned.Started
			exists.Ended = scanned.Ended //|| scanned.cleared
			exists.EndMessage = scanned.EndMessage
		}

		_, err := s.updateNextSessionTmp(scanned)
		if err != nil {
			return err
		}
	}

	return
}

// Attempt to elect a session if no session is currently elected.
// Lower priority number is higher priority
func (s *screenTailer) electSession() (err error) {
	if s.electedSession == nil {
		// FIXME: do not tail notifications here !
		// 1- If no elected session, firstly print notifications
		// err = s.tailNotifications()
		// if err != nil {
		// 	return err
		// }

		// 2- Scan serialized session
		err = s.scanSessions()
		if err != nil {
			return err
		}

		for s.electedSession == nil && s.blockingSessionsQueue.Len() > 0 {
			// 3a- Dequeue next session to tail
			sessionName := s.blockingSessionsQueue.Front()
			if session, ok := s.sessions[*sessionName]; ok {
				s.electedSession = session
			}
			// Remove item
			s.blockingSessionsQueue.PopFront()
		}

		if s.electedSession == nil {
			// 3b- Elect a new session to tail
			priorities := collectionz.Keys(s.sessionsByPriority)
			slices.Sort(priorities)
		end:
			for _, priority := range priorities {
				sessions, ok := s.sessionsByPriority[priority]
				if ok {
					for _, session := range sessions {
						if !session.Started || session.Ended && session.flushed {
							continue
						}
						s.electedSession = session
						break end
					}
				}
			}
		}

		if s.electedSession != nil {
			logger.Debug("elected new session", "electedSession", s.electedSession.Name)
		}
	} else if !s.electedSession.cleared {
		path := sessionSerializedPath(filepath.Dir(s.electedSession.TmpPath), s.electedSession.Name)
		err = updateSession(s.electedSession, path)
		if err != nil {
			return err
		}
	}

	return
}

func (s *screenTailer) tailNotifications() error {
	n, err := filez.CopyChunk(s.notifier.tmpOut, s.outputs.Out(), buf, s.notifier.cursorOut, -1)
	if err != nil {
		return fmt.Errorf("error tailing notifier out: %w", err)
	}
	s.notifier.cursorOut += int64(n)
	n, err = filez.CopyChunk(s.notifier.tmpErr, s.outputs.Err(), buf, s.notifier.cursorErr, -1)
	if err != nil {
		return fmt.Errorf("error tailing notifier err: %w", err)
	}
	s.notifier.cursorErr += int64(n)
	err = s.outputs.Flush()
	return err
}

func (s *screenTailer) tailSession(session *session) (err error) {
	if session.Ended && session.flushed && session.tailed {
		return
	}
	session.tailed = true

	// TODO LOOP :
	// TODO: check if next session tmp available
	// Tail current session tmp
	// if next session tmp available before tailing loop : update next session tmp and tail new session tmp.

	hasNextBefore, _, _ := hasNextSessionTmp(session)
	for {
		//fmt.Printf("\n<<>> Copying file: %s from %d | %d ...\n", session.tmpErr.Name(), session.cursorOut, session.cursorErr)

		// Copy tmp files into outputs
		n, err := filez.CopyChunk(session.tmpOut, s.outputs.Out(), buf, session.cursorOut, -1)
		if err != nil {
			return fmt.Errorf("error tailing session %s out: %w", session.Name, err)
		}
		session.cursorOut += int64(n)
		logger.Debug("flushing session out ...", "session", session.Name, "tmpOut", session.tmpOut.Name(), "n", n, "cursorOut", session.cursorOut)

		n, err = filez.CopyChunk(session.tmpErr, s.outputs.Err(), buf, session.cursorErr, -1)
		if err != nil {
			return fmt.Errorf("error tailing session %s err: %w", session.Name, err)
		}
		session.cursorErr += int64(n)
		logger.Debug("flushing session err ...", "session", session.Name, "tmpErr", session.tmpErr.Name(), "n", n, "cursorErr", session.cursorErr)

		// Check if new session tmp files exists
		//hasNextAfter, err := s.updateNextSessionTmp(session)
		// if err != nil {
		// 	return err
		// }

		hasNextAfter, _, _ := hasNextSessionTmp(session)
		if hasNextAfter && !hasNextBefore {
			// Could have miss some chunk in current tmp files
			hasNextBefore = true
			continue
		} else if hasNextAfter {
			updated, err := s.updateNextSessionTmp(session)
			if err != nil {
				return err
			}
			if !updated {
				break
			}
		} else {
			break
		}
	}

	if session.Ended {
		session.flushed = true
		err = session.tmpOut.Close()
		if err != nil {
			return fmt.Errorf("error closing session %s out: %w", session.Name, err)
		}
		err = session.tmpErr.Close()
		if err != nil {
			return fmt.Errorf("error closing session %s err: %w", session.Name, err)
		}
		logger.Debug("end flushing closed session", "session", session.Name)
	}
	return
}

// Attempt to elect a session, tail elected session, and tail notifications if session is not opened.
func (s *screenTailer) tailNext() (eneded bool, err error) {
	pt := logger.PerfTimer("tmpPath", s.tmpPath, "electedSession", s.electedSession, "blockingSessionsQueueLen", s.blockingSessionsQueue.Len())
	defer pt.End()

	err = s.electSession()
	if err != nil {
		return
	}

	if s.electedSession == nil || !s.electedSession.tailed {
		err = s.tailNotifications()
	}
	// if s.electedSession == nil || s.electedSession.cursorErr == 0 && s.electedSession.cursorOut == 0 {
	// 	err = s.tailNotifications()
	// 	if err != nil {
	// 		return
	// 	}
	// }

	if s.electedSession == nil {
		return
	}

	err = s.tailSession(s.electedSession)
	if err != nil {
		return
	}

	if s.electedSession.Ended {
		eneded = true
		s.electedSession = nil
		err = s.tailNotifications()
		return
	}

	return
}

// tail everithing possible (notifications and sessions already ended)
func (s *screenTailer) tailAll() (err error) {
	ended := true
	for ended {
		ended, err = s.tailNext()
		if err != nil {
			return
		}
	}
	return
}

func clearSessionsMap(sessions *map[string]*session, name string) error {
	if session, ok := (*sessions)[name]; ok {
		err := session.clear()
		if err != nil {
			return err
		}
		delete(*sessions, name)
	}
	return nil
}

// FIXME: does a sync screen exists ?
func NewScreen(outputs printz.Outputs) *screen {
	// FIXME: not implemented
	panic("not implemented yet")
}

func NewAsyncScreen(tmpPath string, force bool) *screen {
	//if _, err := os.Stat(tmpPath); err == nil {
	//	panic(fmt.Sprintf("unable to create async screen: [%s] path already exists", tmpPath))
	//}

	err := os.MkdirAll(tmpPath, tmpDirFileMode)
	if err != nil {
		panic(err)
	}

	screenLockFilepath := filepath.Join(tmpPath, screenLockFilename)

	if force {
		// On force, override zcreen file locking
		ok, err := filez.Exists(screenLockFilepath)
		if err != nil {
			panic(fmt.Errorf("Unable to stat file: %s", screenLockFilepath))
		}
		if ok {
			// Lock file already exists

			screenLock := flock.New(screenLockFilepath)
			// Always unlock on start to acquire the lock
			err = screenLock.Close()
			if err != nil {
				panic(fmt.Errorf("unable to acquire the lock on dir: %s ; err: %w", tmpPath, err))
			}
			err = os.Remove(screenLockFilepath)
			if err != nil {
				panic(fmt.Errorf("unable to remove lock on dir: %s ; err: %w", tmpPath, err))
			}
		}
	}

	screenLock := flock.New(screenLockFilepath)
	// err = utilz.FileLock(screenLock, time.Second)
	// if err != nil {
	// 	panic(fmt.Errorf("unable to create a new zcreen, dir: %s is lock by another instance: %w", tmpPath, err))
	// }

	lockFilepath := filepath.Join(tmpPath, lockFilename)
	return &screen{
		tmpPath:    tmpPath,
		screenLock: screenLock,
		fileLock:   flock.New(lockFilepath), // FIXME: rename syncLock
		sessions:   make(map[string]*session),
		notifier:   buildPrinter(tmpPath, notifierPrinterName, 0),
	}
}

func Clear(zcreenPath string) error {
	err := os.RemoveAll(zcreenPath)
	return err
}

func ClearSession(zcreenPath, name string) error {
	err := clearSessionFiles(zcreenPath, name)
	return err
}
