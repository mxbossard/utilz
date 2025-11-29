package zcreen

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/mxbossard/utilz/collectionz"
	"github.com/mxbossard/utilz/errorz"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/printz"
	"github.com/mxbossard/utilz/utilz"
)

type screenTailer struct {
	sync.Mutex
	fileLock              *flock.Flock
	tmpPath               string
	outputs               printz.Outputs
	electedSession        *session
	sessions              map[string]*session
	sessionsByPriority    map[int][]*session
	notifier              *printer
	blockingSessionsQueue *collectionz.Queue[string]
}

func (s *screenTailer) tailOnce(sessionName string) (tailed, ended bool, err error) {
	err = s.scanSessions()
	if err != nil {
		return
	}

	blocking := s.sessions[sessionName]
	if blocking == nil {
		return
	}

	if blocking.cleared {
		// FIXME: probably not used anymore
		err = s.ClearSession(sessionName)
		return
	}

	logger.Debug("tailing ...", "session", sessionName)

	err = blocking.Flush()
	if err != nil {
		return
	}

	err = s.tailSession(blocking)
	if err != nil {
		return
	}

	// tailed = blocking.cursorOut > 0 || blocking.cursorErr > 0
	tailed = blocking.tailed
	ended = blocking.Ended

	// path := sessionSerializedPath(filepath.Dir(blocking.TmpPath), blocking.Name)
	// err = updateSession(blocking, path)
	// if err != nil {
	// 	return err
	// }

	return
}

// Tail continuously until session is ended.
// Tail only supplied session
// If session does not exists, do nothing.
func (s *screenTailer) TailOnlyBlocking(sessionName string, timeout time.Duration) error {
	pt := logger.PerfTimer("sessionName", sessionName)
	defer pt.End()

	var blocking *session
	startTime := time.Now()

	for blocking == nil || !blocking.Ended {
		if time.Since(startTime) > timeout {
			err := errorz.Timeoutf(timeout, "TailOnlyBlocking() for session: [%s]", sessionName)
			return err
		}

		_, _, err := s.tailOnce(sessionName)
		if err != nil {
			return err
		}

		blocking = s.sessions[sessionName]
		if blocking == nil || !blocking.Ended {
			time.Sleep(continuousFlushPeriod)

			err = s.scanSessions()
			if err != nil {
				return err
			}
		}

		if s.sessions[sessionName] == nil {
			// Session does not exist after a scan
			return nil
		}

	}

	path := sessionSerializedPath(filepath.Dir(blocking.TmpPath), blocking.Name)
	err := updateSession(blocking, path)
	if err != nil {
		return err
	}

	return nil
}

// Tail continuously until session is ended.
// Put supplied session on top for next session election.
func (s *screenTailer) TailBlocking(sessionName string, timeout time.Duration) error {
	pt := logger.PerfTimer("sessionName", sessionName)
	defer pt.End()

	var blocking *session
	startTime := time.Now()
	// Push session on top of priority queue
	s.blockingSessionsQueue.PushFront(sessionName)
	// Wait and find session

	for blocking = s.sessions[sessionName]; blocking == nil || !blocking.Ended; {
		if time.Since(startTime) > timeout {
			var err error
			if blocking != nil {
				err = errorz.Timeoutf(timeout, "TailBlocking() for session: [%s] (cursorOut: %d ; cursorErr: %d ; end: %v ; endMsg: %s ; flush: %v ; tailed: %v)", sessionName, blocking.cursorOut, blocking.cursorErr, blocking.Ended, blocking.EndMessage, blocking.flushed, blocking.tailed)
			} else {
				err = errorz.Timeoutf(timeout, "TailBlocking() for session: [%s]", sessionName)
			}
			return err
		}

		logger.Debug("tailing ...", "session", sessionName)
		err := s.tailAll()
		if err != nil {
			return err
		}

		blocking = s.sessions[sessionName]
		if blocking == nil || !blocking.Ended {
			time.Sleep(continuousFlushPeriod)

			err = s.scanSessions()
			if err != nil {
				return err
			}
		}
	}

	if blocking == nil || blocking.Ended {
		// FIXME: could be called in loop ^^ if tailAll() managed all cases
		err := s.tailNotifications()
		if err != nil {
			return err
		}
	}

	return nil
}

// Tail continuously until all supplied sessions are ended.
func (s *screenTailer) TailSuppliedBlocking0(sessionNames []string, timeout time.Duration) error {
	pt := logger.PerfTimer()
	defer pt.End()

	startTime := time.Now()

	// 1- Priorize supplied sessions
	for k := len(sessionNames) - 1; k >= 0; k-- {
		session := sessionNames[k]
		s.blockingSessionsQueue.PushFront(session)
	}

	// fmt.Printf("\n<<>> blockingSessionsQueue: %s\n", s.blockingSessionsQueue)

	_, err := s.tailNext()
	if err != nil {
		return err
	}

	var notEnded []*session
	for notEnded == nil || len(notEnded) > 0 {
		notEndedNames := collectionz.Map(&notEnded, func(s *session) string { return s.Name })
		if time.Since(startTime) > timeout {
			err := errorz.Timeoutf(timeout, "TailSuppliedBlocking(), some sessions not ended after timeout: %s", notEndedNames)
			return err
		}

		// Updating not ended session list
		notEnded = nil
		allSessions := collectionz.Values(s.sessions)
		for _, s := range allSessions {
			if collectionz.Contains(&sessionNames, s.Name) {
				notEnded = append(notEnded, s)
			}
		}

		var ended []int
		for pos, s := range notEnded {
			if s.Ended {
				ended = append(ended, pos)
			}
		}
		removed := 0
		for _, pos := range ended {
			notEnded = collectionz.RemoveFast(notEnded, pos-removed)
			removed++
		}

		if len(notEnded) > 0 {
			time.Sleep(continuousFlushPeriod)
			err = s.scanSessions()
			if err != nil {
				return err
			}
		}
		logger.Debug("TailSuppliedBlocking tailNext ...", "notEnded", notEnded, "ended", ended)
		_, err := s.tailNext()
		if err != nil {
			return err
		}
	}

	return nil
}

// Tail continuously until all supplied sessions are ended.
func (s *screenTailer) TailSuppliedBlocking(sessionNames []string, timeout time.Duration) error {
	pt := logger.PerfTimer()
	defer pt.End()

	startTime := time.Now()

	err := s.tailNotifications()
	if err != nil {
		return err
	}

	for _, session := range sessionNames {
		var tailed, ended bool
		// k := 0
		for !ended {
			if time.Since(startTime) > timeout {
				err := errorz.Timeoutf(timeout, "TailSuppliedBlocking() for session: [%s]", sessionNames)
				return err
			}

			if !tailed {
				// fmt.Printf("\n<<>> tailing notif before session: %s #%d\n", session, k)
				// k++
				// fmt.Printf("\n<<>> notifier %s / %s ; cursorOut: %d \n", session, s.electedSession, s.notifier.cursorOut)
				err = s.tailNotifications()
				if err != nil {
					return err
				}
			}

			tailed, ended, err = s.tailOnce(session)
			if err != nil {
				return err
			}

			if !ended {
				time.Sleep(continuousFlushPeriod)
			}
		}

		fmt.Printf("session %s is ended\n", session)

		blocking := s.sessions[session]
		if blocking != nil {
			path := sessionSerializedPath(filepath.Dir(blocking.TmpPath), blocking.Name)
			err := updateSession(blocking, path)
			if err != nil {
				return err
			}
		}
	}

	err = s.tailNotifications()
	if err != nil {
		return err
	}

	return nil
}

// Flush continuously until all sessions are ended.
func (s *screenTailer) TailAllBlocking(timeout time.Duration) error {
	pt := logger.PerfTimer()
	defer pt.End()

	startTime := time.Now()

	err := s.tailAll()
	if err != nil {
		return err
	}

	var notEnded []*session
	for notEnded == nil || len(notEnded) > 0 {
		notEndedNames := collectionz.Map(&notEnded, func(s *session) string { return s.Name })
		if time.Since(startTime) > timeout {
			err := errorz.Timeoutf(timeout, "TailAllBlocking(), some sessions not ended after timeout: %s", notEndedNames)
			return err
		}

		// Updating not ended session list
		notEnded = collectionz.Values(s.sessions)
		var ended []int
		for pos, s := range notEnded {
			if s.Ended {
				ended = append(ended, pos)
			}
		}
		removed := 0
		for _, pos := range ended {
			notEnded = collectionz.RemoveFast(notEnded, pos-removed)
			removed++
		}

		if len(notEnded) > 0 {
			time.Sleep(continuousFlushPeriod)
			err = s.scanSessions()
			if err != nil {
				return err
			}
		}
		logger.Debug("TailAllBlocking flushAll ...", "notEnded", notEnded, "ended", ended)
		err := s.tailAll()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *screenTailer) Reclaim(session string) error {
	// TODO
	panic("not implemented yet")
}

func (s *screenTailer) ReclaimAll() error {
	// TODO
	panic("not implemented yet")
}

func (s *screenTailer) clearSession(name string) error {
	//fmt.Printf("Clearing tailer session dir: [%s] ...\n", name)
	for p, sessions := range s.sessionsByPriority {
		for _, session := range sessions {
			if session.Name == name {
				sessions = collectionz.Delete(sessions, session)
				s.sessionsByPriority[p] = sessions
			}
		}
	}

	if s.electedSession != nil && s.electedSession.Name == name {
		s.electedSession = nil
	}

	err := clearSessionFiles(s.tmpPath, name)
	if err != nil {
		return err
	}
	logger.Debug("Tailer: cleared session files", "name", name)
	err = clearSessionsMap(&s.sessions, name)
	if err != nil {
		return err
	}
	logger.Debug("Tailer: cleared session map", "name", name)
	return err
}

func (s *screenTailer) ClearSession(name string) error {
	logger.Debug("Tailer: clearing session ...", "name", name)
	s.Lock()
	defer s.Unlock()
	err := utilz.FileLock(s.fileLock, fileLockingTimeout)
	if err != nil {
		return err
	}
	defer utilz.FileUnlock(s.fileLock)

	return s.clearSession(name)
}

func (s *screenTailer) Clear() (err error) {
	s.Lock()
	defer s.Unlock()
	err = utilz.FileLock(s.fileLock, fileLockingTimeout)
	if err != nil {
		return
	}
	defer utilz.FileUnlock(s.fileLock)

	sessions := collectionz.Keys(s.sessions)
	for _, session := range sessions {
		err := s.clearSession(session)
		if err != nil {
			return err
		}
	}
	return
}

func NewAsyncScreenTailer(outputs printz.Outputs, tmpPath string) *screenTailer {
	if ok, _ := filez.IsDirectory(tmpPath); !ok {
		panic(fmt.Sprintf("unable to create read only async screen tailer: [%s] path do not exists", tmpPath))
	}

	notifier := buildReadOnlyPrinter(tmpPath, notifierPrinterName, 0)
	lockFilepath := filepath.Join(tmpPath, lockFilename)
	s := &screenTailer{
		outputs:               outputs,
		tmpPath:               tmpPath,
		fileLock:              flock.New(lockFilepath),
		sessions:              make(map[string]*session),
		sessionsByPriority:    make(map[int][]*session),
		notifier:              notifier,
		blockingSessionsQueue: collectionz.NewQueue[string](),
	}

	return s
}

func NewAsyncScreenTailerWaiting(outputs printz.Outputs, tmpPath string, timeout time.Duration) *screenTailer {
	startTime := time.Now()
	for ok, err := filez.IsDirectory(tmpPath); !ok; {
		if err != nil {
			panic(err)
		}
		if time.Since(startTime) > timeout {
			panic(fmt.Sprintf("unable to create read only async screen tailer: [%s] path do not exists after timeout: [%s]", tmpPath, timeout))
		}
		time.Sleep(10 * time.Millisecond)
		ok, err = filez.IsDirectory(tmpPath)
	}
	s := NewAsyncScreenTailer(outputs, tmpPath)
	return s
}
