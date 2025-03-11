package screen

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"mby.fr/utils/collections"
	"mby.fr/utils/errorz"
	"mby.fr/utils/filez"
	"mby.fr/utils/printz"
	"mby.fr/utils/zlog"
)

const (
	tmpDirFileMode        = 0760
	notifierPrinterName   = "_-_notifier"
	continuousFlushPeriod = 1 * time.Millisecond
)

var (
	buf    = make([]byte, bufLen)
	logger = zlog.New()
)

type screen struct {
	sync.Mutex
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
	s.sessions[name] = session
	return session
}

func (s *screen) ClearSession(name string) error {
	s.Lock()
	defer s.Unlock()
	//fmt.Printf("Clearing sink session dir: [%s] ...\n", name)
	return clearSession(&s.sessions, name)
}

func (s *screen) NotifyPrinter() printz.Printer {
	return s.notifier.Printer
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
				allSessions := collections.Values(s.sessions)
				notEndedNames := collections.Map(&allSessions, func(s *session) string { return s.Name })
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

func (*screen) ConfigPrinter(name string) (s *session) {
	// TODO later
	return
}

type screenTailer struct {
	sync.Mutex
	tmpPath               string
	outputs               printz.Outputs
	electedSession        *session
	sessions              map[string]*session
	sessionsByPriority    map[int][]*session
	notifier              *printer
	blockingSessionsQueue *collections.Queue[string]
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
		notEndedNames := collections.Map(&notEnded, func(s *session) string { return s.Name })
		//fmt.Printf("Flushing sessions: %v\n", notEndedNames)
		if time.Since(startTime) > timeout {
			err := errorz.Timeoutf(timeout, "TailAllBlocking(), some sessions not ended after timeout: %s", notEndedNames)
			return err
		}

		// Updating not ended session list
		notEnded = collections.Values(s.sessions)
		var ended []int
		for pos, s := range notEnded {
			if s.Ended {
				ended = append(ended, pos)
			}
		}
		removed := 0
		for _, pos := range ended {
			notEnded = collections.RemoveFast(notEnded, pos-removed)
			removed++
		}

		if len(notEnded) > 0 {
			time.Sleep(continuousFlushPeriod)
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

func (s *screenTailer) ClearSession(name string) error {
	s.Lock()
	defer s.Unlock()
	//fmt.Printf("Clearing tailer session dir: [%s] ...\n", name)
	return clearSession(&s.sessions, name)
}

func (s *screenTailer) Clear() (err error) {
	sessions := collections.Keys(s.sessions)
	for _, session := range sessions {
		err := s.ClearSession(session)
		if err != nil {
			return err
		}
	}
	return
}

func (s *screenTailer) scanSessions() (err error) {
	pt := logger.PerfTimer("tmpPath", s.tmpPath)
	defer pt.End("sessionsCount", len(s.sessions))

	wildcard := sessionSerializedPath(s.tmpPath, "*")
	sers, err := filepath.Glob(wildcard)
	if err != nil {
		return err
	}

	for _, filePath := range sers {
		scanned, err := deserializeSession(filePath)
		if err != nil {
			return err
		}

		// clear session
		scanned.currentPriority = nil
		scanned.tmpOut = nil
		scanned.tmpErr = nil
		if exists, ok := s.sessions[scanned.Name]; !ok {
			// init session
			scanned.tmpOut, err = os.OpenFile(scanned.TmpOutName, os.O_RDONLY, 0)
			if err != nil {
				return fmt.Errorf("error opening session out tmp file (%s): %w", scanned.TmpOutName, err)
			}
			scanned.tmpErr, err = os.OpenFile(scanned.TmpErrName, os.O_RDONLY, 0)
			if err != nil {
				return fmt.Errorf("error opening session err tmp file (%s): %w", scanned.TmpErrName, err)
			}
			s.sessions[scanned.Name] = scanned
			s.sessionsByPriority[scanned.PriorityOrder] = append(s.sessionsByPriority[scanned.PriorityOrder], scanned)
		} else {
			// update session
			exists.Started = scanned.Started
			exists.Ended = scanned.Ended

		}
	}
	pt.End("sessionsCount", len(s.sessions), "sers", sers)
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

		if s.blockingSessionsQueue.Len() > 0 {
			// 3a- Dequeue next session to tail
			sessionName := s.blockingSessionsQueue.Front()
			//fmt.Printf("===== dequeuing prioritary session: [%s]\n", *sessionName)
			if session, ok := s.sessions[*sessionName]; ok {
				s.electedSession = session
				// Remove item
				s.blockingSessionsQueue.PopFront()
			} else {
				//fmt.Printf("=== ERROR session: [%s] not found\n", *sessionName)
				//err = fmt.Errorf("session with name: [%s] not found", sessionName)
				//return err
			}
		}

		if s.electedSession == nil {
			// 3b- Elect a new session to tail
			priorities := collections.Keys(s.sessionsByPriority)
			slices.Sort(priorities)
		end:
			for _, priority := range priorities {
				sessions, ok := s.sessionsByPriority[priority]
				if ok {
					for _, session := range sessions {
						//fmt.Printf("electing session #%d: [%s] ; ended: %v ; flushed: %v\n", k, session.Name, session.Ended, session.flushed)
						if !session.Started || session.Ended && session.flushed {
							continue
						}
						//fmt.Printf("elected session: [%s]\n", session.Name)
						s.electedSession = session
						break end
					}
				}
			}
		}

		if s.electedSession != nil {
			logger.Debug("elected new session", "electedSession", s.electedSession.Name)
		}
	} else {
		//path := serializedPath(s.electedSession)
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

func clearSession(sessions *map[string]*session, name string) error {
	if session, ok := (*sessions)[name]; ok {
		err := session.Clear()
		if err != nil {
			return err
		}
		delete(*sessions, name)
	} else {
		//return fmt.Errorf("session: [%s] does not exists", name)
	}
	return nil
}

func NewScreen(outputs printz.Outputs) *screen {
	// FIXME: not implemented
	return &screen{
		sessions: make(map[string]*session),
	}
}

func NewAsyncScreen(tmpPath string) *screen {
	if _, err := os.Stat(tmpPath); err == nil {
		panic(fmt.Sprintf("unable to create async screen: [%s] path already exists", tmpPath))
	}

	err := os.MkdirAll(tmpPath, tmpDirFileMode)
	if err != nil {
		panic(err)
	}

	return &screen{
		tmpPath:  tmpPath,
		sessions: make(map[string]*session),
		notifier: buildPrinter(tmpPath, notifierPrinterName, 0),
	}
}

func NewAsyncScreenTailer(outputs printz.Outputs, tmpPath string) *screenTailer {
	if ok, _ := filez.IsDirectory(tmpPath); !ok {
		panic(fmt.Sprintf("unable to create read only async screen: [%s] path do not exists", tmpPath))
	}

	notifier := buildPrinter(tmpPath, notifierPrinterName, 0)
	return &screenTailer{
		outputs:               outputs,
		tmpPath:               tmpPath,
		sessions:              make(map[string]*session),
		sessionsByPriority:    make(map[int][]*session),
		notifier:              notifier,
		blockingSessionsQueue: collections.NewQueue[string](),
	}
}

func NewAsyncScreenTailerWaiting(outputs printz.Outputs, tmpPath string, timeout time.Duration) *screenTailer {
	startTime := time.Now()
	for ok, _ := filez.IsDirectory(tmpPath); !ok; {
		if time.Since(startTime) > timeout {
			panic(fmt.Sprintf("unable to create read only async screen: [%s] path do not exists after timeout: [%s]", tmpPath, timeout))
		}
		time.Sleep(1 * time.Millisecond)
	}
	return NewAsyncScreenTailer(outputs, tmpPath)
}
