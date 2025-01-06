package screen

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"mby.fr/utils/collections"
	"mby.fr/utils/filez"
	"mby.fr/utils/printz"
	"mby.fr/utils/zlog"
)

const (
	notifierPrinterName   = "notifier"
	continuousFlushPeriod = 1 * time.Millisecond
)

var (
	buf    = make([]byte, bufLen)
	logger = zlog.New()
)

type Sink interface {
	Session(string, int) *session
	NotifyPrinter() printz.Printer
	//Flush() error
	Close() error
	FlushBlocking(string, time.Duration) error
	FlushAllBlocking(time.Duration) error
}

type Session interface {
	Start(time.Duration) error
	End() error
	Flush() error
	Printer(string, int) printz.Printer
	ClosePrinter(string) error
}

type Tailer interface {
	//Flush() error
	TailBlocking(string, time.Duration) error
	TailAllBlocking(time.Duration) error
}

type screen struct {
	sync.Mutex
	tmpPath  string
	sessions map[string]*session
	notifier *printer
}

func (s *screen) Session(name string, priorityOrder int) *session {
	s.Lock()
	defer s.Unlock()
	if session, ok := s.sessions[name]; ok {
		return session
	}

	session := buildSession(name, priorityOrder, s.tmpPath)
	s.sessions[name] = session
	return session
}

func (s *screen) NotifyPrinter() printz.Printer {
	return s.notifier.Printer
}

func (s *screen) FlushBlocking(sessionName string, timeout time.Duration) (err error) {
	s.Lock()
	defer s.Unlock()
	startTime := time.Now()
	if session, ok := s.sessions[sessionName]; ok {
		for !session.Ended {
			if time.Since(startTime) > timeout {
				err := fmt.Errorf("timeout FlushBlocking() for session: [%s]", sessionName)
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
				err := fmt.Errorf("timeout FlushAllBlocking() some sessions: [%s]", notEndedNames)
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

func (*screen) ConfigPrinter(name string) (s *session) {
	// TODO later
	return
}

type screenTailer struct {
	tmpPath               string
	outputs               printz.Outputs
	electedSession        *session
	sessions              map[string]*session
	sessionsByPriority    map[int][]*session
	notifier              *printer
	blockingSessionsQueue *collections.Queue[string]
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

	return nil
}

func (s *screenTailer) scanSessions() (err error) {
	sers, err := filepath.Glob(s.tmpPath + "/*.ser")
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
				return err
			}
			scanned.tmpErr, err = os.OpenFile(scanned.TmpErrName, os.O_RDONLY, 0)
			if err != nil {
				return err
			}
			s.sessions[scanned.Name] = scanned
			s.sessionsByPriority[scanned.PriorityOrder] = append(s.sessionsByPriority[scanned.PriorityOrder], scanned)
		} else {
			// update session
			exists.Started = scanned.Started
			exists.Ended = scanned.Ended

		}
	}
	return
}

func (s *screenTailer) flush() (err error) {
	// Flush the display : print sessions in order onto std outputs (keep written bytes count)

	if s.electedSession == nil {
		// 1- If no elected session, firstly print notifications
		n, err := filez.CopyChunk(s.notifier.tmpOut, s.outputs.Out(), buf, s.notifier.cursorOut, -1)
		if err != nil {
			return err
		}
		s.notifier.cursorOut += int64(n)
		n, err = filez.CopyChunk(s.notifier.tmpErr, s.outputs.Err(), buf, s.notifier.cursorErr, -1)
		if err != nil {
			return err
		}
		s.notifier.cursorErr += int64(n)

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
	} else {
		path := serializedPath(s.electedSession)
		err = updateSession(s.electedSession, path)
		if err != nil {
			return err
		}
	}

	if s.electedSession == nil {
		return
	}
	//fmt.Printf("flushtailing session: [%s](%s) ; ended: %v ; flushed: %v\n", s.electedSession.Name, s.electedSession.tmpOut.Name(), s.electedSession.Ended, s.electedSession.flushed)

	n, err := filez.CopyChunk(s.electedSession.tmpOut, s.outputs.Out(), buf, s.electedSession.cursorOut, -1)
	if err != nil {
		return err
	}
	s.electedSession.cursorOut += int64(n)

	n, err = filez.CopyChunk(s.electedSession.tmpErr, s.outputs.Err(), buf, s.electedSession.cursorErr, -1)
	if err != nil {
		return err
	}
	s.electedSession.cursorErr += int64(n)

	if s.electedSession.Ended {
		s.electedSession.flushed = true
		err = s.electedSession.tmpOut.Close()
		if err != nil {
			return err
		}
		err = s.electedSession.tmpErr.Close()
		if err != nil {
			return err
		}
		s.electedSession = nil
		// FIXME: do not use a recusrsive call
		s.flush()
	}

	return
}

/** Flush continuously until session is ended. */
/** Put supplied session on top for next session election. */
func (s *screenTailer) TailBlocking(sessionName string, timeout time.Duration) error {
	var blocking *session
	startTime := time.Now()
	// Push session on top of priority queue
	s.blockingSessionsQueue.PushFront(sessionName)
	// Wait and find session

	for blocking = s.sessions[sessionName]; blocking == nil || !blocking.Ended; {
		if time.Since(startTime) > timeout {
			err := fmt.Errorf("timeout TailBlocking() for session: [%s]", sessionName)
			return err
		}

		err := s.flush()
		if err != nil {
			return err
		}

		blocking = s.sessions[sessionName]
		if blocking == nil || !blocking.Ended {
			time.Sleep(continuousFlushPeriod)
		}

		if blocking != nil {
			//fmt.Printf("Waited for session: [%s] (ended: %v)\n", sessionName, blocking.Ended)
		} else {
			//fmt.Printf("Waited for session: [%s]\n", sessionName)
		}
	}

	return nil
}

/** Flush continuously until all sessions are ended. */
func (s *screenTailer) TailAllBlocking(timeout time.Duration) error {
	startTime := time.Now()

	err := s.flush()
	if err != nil {
		return err
	}

	var notEnded []*session
	for len(notEnded) > 0 {
		notEndedNames := collections.Map(&notEnded, func(s *session) string { return s.Name })
		//fmt.Printf("Flushing sessions: %v\n", notEndedNames)
		if time.Since(startTime) > timeout {
			err := fmt.Errorf("timeout TailAllBlocking(), some sessions not ended after timeout: %s", notEndedNames)
			return err
		}

		notEnded := collections.Values(s.sessions)

		var ended []int
		for pos, s := range notEnded {
			if s.Ended {
				ended = append(ended, pos)
			}
		}

		for _, pos := range ended {
			collections.RemoveFast(notEnded, pos)
		}

		if len(notEnded) > 0 {
			time.Sleep(continuousFlushPeriod)
		}
		err := s.flush()
		if err != nil {
			return err
		}
	}

	return nil
}

func (*screenTailer) AsyncFlush0(timeout time.Duration) (err error) {
	// Launch goroutine wich will continuously flush async display
	// TODO later ?
	return
}

func (*screenTailer) BlockTail0(timeout time.Duration) (err error) {
	// Tail async display blocking until end
	// TODO later ?
	return
}

func NewScreen(outputs printz.Outputs) *screen {
	return &screen{
		sessions: make(map[string]*session),
		//outputs:  printz.NewStandardOutputs(),
	}
}

func NewAsyncScreen(tmpPath string) *screen {
	if _, err := os.Stat(tmpPath); err == nil {
		panic(fmt.Sprintf("unable to create async screen: [%s] path already exists", tmpPath))
	}

	err := os.MkdirAll(tmpPath, 0760)
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
