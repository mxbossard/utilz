package zcreen

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	fileLock *flock.Flock
	tmpPath  string
	sessions map[string]*session
	notifier *printer
}

func (s *screen) Session(name string, priorityOrder int) *session {
	if name == "" {
		panic("cannot get session of empty name")
	}
	s.Lock()
	defer s.Unlock()
	if session, ok := s.sessions[name]; ok {
		return session
	}

	session := buildSession(name, priorityOrder, s.tmpPath)
	//fmt.Printf("Built sink session: [%s]\n", name)
	s.sessions[name] = session
	return session
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
	err = s.notifier.tmpOut.Close()
	if err != nil {
		return err
	}
	err = s.notifier.tmpErr.Close()
	return
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

// Flush continuously until session is ended.
// Tail only supplied session
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

		err := s.scanSessions()
		if err != nil {
			return err
		}

		blocking = s.sessions[sessionName]
		if blocking != nil && blocking.cleared {
			err = s.ClearSession(sessionName)
			return err
		}
		logger.Debug("tailing ...", "session", sessionName)
		if blocking != nil {
			err := s.tailSession(blocking)
			if err != nil {
				return err
			}
		}

		if blocking == nil || !blocking.Ended {
			time.Sleep(continuousFlushPeriod)
		}
	}

	path := sessionSerializedPath(filepath.Dir(blocking.TmpPath), blocking.Name)
	err := updateSession(blocking, path)
	if err != nil {
		return err
	}

	return nil
}

// Flush continuously until session is ended.
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
			err := errorz.Timeoutf(timeout, "TailBlocking() for session: [%s]", sessionName)
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

// Flush continuously until all supplied sessions are ended.
func (s *screenTailer) TailSuppliedBlocking(sessionNames []string, timeout time.Duration) error {
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
		logger.Debug("TailAllBlocking flushAll ...", "notEnded", notEnded, "ended", ended)
		err := s.tailAll()
		if err != nil {
			return err
		}
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
		sessions = append(sessions, scanned)
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
		if scanned.cleared {
			// Remove cleared session
			err = os.Remove(scanned.serializationFilepath)
			if err != nil {
				return fmt.Errorf("unable to remove serialized session file (%s): %w", scanned.serializationFilepath, err)
			}
		}

		// clear session
		scanned.currentPriority = nil
		scanned.tmpOut = nil
		scanned.tmpErr = nil
		if exists, ok := s.sessions[scanned.Name]; !ok {
			// init session
			if !scanned.cleared {
				scanned.tmpOut, err = os.OpenFile(scanned.TmpOutName, os.O_RDONLY, 0)
				if err != nil {
					return fmt.Errorf("error opening session out tmp file (%s): %w", scanned.TmpOutName, err)
				}
				scanned.tmpErr, err = os.OpenFile(scanned.TmpErrName, os.O_RDONLY, 0)
				if err != nil {
					return fmt.Errorf("error opening session err tmp file (%s): %w", scanned.TmpErrName, err)
				}
			}
			scanned.Ended = scanned.Ended || scanned.cleared
			s.sessions[scanned.Name] = scanned
			s.sessionsByPriority[scanned.PriorityOrder] = append(s.sessionsByPriority[scanned.PriorityOrder], scanned)
		} else {
			// update session
			exists.Started = scanned.Started
			exists.Ended = scanned.Ended || scanned.cleared
			exists.cleared = scanned.cleared

			if !exists.cleared {
				if exists.tmpOut == nil && scanned.TmpOutName != "" {
					exists.tmpOut, err = os.OpenFile(scanned.TmpOutName, os.O_RDONLY, 0)
					if err != nil {
						return fmt.Errorf("error opening session out tmp file (%s): %w", scanned.TmpOutName, err)
					}
				}
				if exists.tmpErr == nil && scanned.TmpErrName != "" {
					exists.tmpErr, err = os.OpenFile(scanned.TmpErrName, os.O_RDONLY, 0)
					if err != nil {
						return fmt.Errorf("error opening session err tmp file (%s): %w", scanned.TmpErrName, err)
					}
				}
			}
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
		err = s.tailNotifications()
		if err != nil {
			return err
		}

		// 2- Scan serialized session
		err = s.scanSessions()
		if err != nil {
			return err
		}

		for s.electedSession == nil && s.blockingSessionsQueue.Len() > 0 {
			//if s.blockingSessionsQueue.Len() > 0 {
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

func NewScreen(outputs printz.Outputs) *screen {
	// FIXME: not implemented
	panic("not implemented yet")
}

func NewAsyncScreen(tmpPath string) *screen {
	if _, err := os.Stat(tmpPath); err == nil {
		panic(fmt.Sprintf("unable to create async screen: [%s] path already exists", tmpPath))
	}

	err := os.MkdirAll(tmpPath, tmpDirFileMode)
	if err != nil {
		panic(err)
	}
	lockFilepath := filepath.Join(tmpPath, lockFilename)
	return &screen{
		tmpPath:  tmpPath,
		fileLock: flock.New(lockFilepath),
		sessions: make(map[string]*session),
		notifier: buildPrinter(tmpPath, notifierPrinterName, 0),
	}
}

func NewAsyncScreenTailer(outputs printz.Outputs, tmpPath string) *screenTailer {
	if ok, _ := filez.IsDirectory(tmpPath); !ok {
		panic(fmt.Sprintf("unable to create read only async screen tailer: [%s] path do not exists", tmpPath))
	}

	notifier := buildPrinter(tmpPath, notifierPrinterName, 0)
	lockFilepath := filepath.Join(tmpPath, lockFilename)
	return &screenTailer{
		outputs:               outputs,
		tmpPath:               tmpPath,
		fileLock:              flock.New(lockFilepath),
		sessions:              make(map[string]*session),
		sessionsByPriority:    make(map[int][]*session),
		notifier:              notifier,
		blockingSessionsQueue: collectionz.NewQueue[string](),
	}
}

func NewAsyncScreenTailerWaiting(outputs printz.Outputs, tmpPath string, timeout time.Duration) *screenTailer {
	startTime := time.Now()
	for ok, _ := filez.IsDirectory(tmpPath); !ok; {
		if time.Since(startTime) > timeout {
			panic(fmt.Sprintf("unable to create read only async screen tailer: [%s] path do not exists after timeout: [%s]", tmpPath, timeout))
		}
		time.Sleep(1 * time.Millisecond)
	}
	return NewAsyncScreenTailer(outputs, tmpPath)
}

func Clear(zcreenPath string) error {
	err := os.RemoveAll(zcreenPath)
	return err
}

func ClearSession(zcreenPath, name string) error {
	err := clearSessionFiles(zcreenPath, name)
	return err
}
