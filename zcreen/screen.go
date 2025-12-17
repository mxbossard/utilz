package zcreen

import (
	"fmt"
	"os"
	"path/filepath"
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
	notifier   *fsPrinter
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
	return s.notifier
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
	var agg errorz.Aggregated

	if s.notifier != nil {
		err = s.notifier.Close("screen closed")
		agg.Add(err)
	}

	for _, s := range s.sessions {
		// Closing screen SHOULD not end sessions but close it : In case of zcreen failure, a restart must take back the session not ended.
		agg.Add(s.close("screen closed"))
	}
	s.closed = true
	return agg.Return()
}

// func (s *screen) clearSession(name string) error {
// 	if session, ok := (s.sessions)[name]; ok {
// 		err := session.clear()
// 		if err != nil {
// 			return err
// 		}
// 		delete(s.sessions, name)
// 	}
// 	return nil
// }

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
		// clearSessionsMap(&s.sessions, sessionName)

		delete(s.sessions, sessionName)
		logger.Debug("Resync: removed session", "session", sessionName)
		// fmt.Printf("Resync: removed session: %s\n", sessionName)
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
	// wildcard := sessionSerializedPath(zcreenPath, "*")
	sers, err := filepath.Glob(filepath.Join(forgeSessionsDirPath(zcreenPath), "*", sessionSerialedFilename))
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
			panic(fmt.Errorf("unable to stat file: %s", screenLockFilepath))
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
		notifier:   buildNotifierPrinter(tmpPath, "", 0),
	}
}

func Clear(zcreenPath string) error {
	err := os.RemoveAll(zcreenPath)
	return err
}

// Does SHOULD be cleared by a tailer ?
func ClearSession(zcreenPath, name string) error {
	err := clearSessionFiles(zcreenPath, name)
	return err
}
