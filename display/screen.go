package display

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"mby.fr/utils/collections"
	"mby.fr/utils/filez"
	"mby.fr/utils/printz"
)

type screen struct {
	tmpPath  string
	sessions map[string]*session
	//sessionsByPriority map[int][]*session
}

func (s *screen) Session(name string, priorityOrder int) *session {
	if session, ok := s.sessions[name]; ok {
		return session
	}

	session := buildSession(name, priorityOrder, s.tmpPath)

	s.sessions[name] = session
	//s.sessionsByPriority[priorityOrder] = append(s.sessionsByPriority[priorityOrder], session)

	return session
}

func (*screen) ConfigPrinter(name string) (s *session) {
	// TODO later
	return
}

type screenTailer struct {
	tmpPath            string
	outputs            printz.Outputs
	electedSession     *session
	sessions           map[string]*session
	sessionsByPriority map[int][]*session
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

func (s *screenTailer) Flush() (err error) {
	// Flush the display : print sessions in order onto std outputs (keep written bytes count)

	if s.electedSession == nil {
		// TODO: scan tmp dir to refresh maps sessions & sessionsByPriority
		// TODO: must build all sessions from files

		sers, err := filepath.Glob(s.tmpPath + "/*.ser")
		if err != nil {
			return err
		}
		for _, filePath := range sers {
			session, err := deserializeSession(filePath)
			if err != nil {
				return err
			}

			// clear session
			session.currentPriority = nil
			session.tmpOut = nil
			session.tmpErr = nil

			if exists, ok := s.sessions[session.Name]; !ok {
				// init session
				session.tmpOut, err = os.OpenFile(session.TmpOutName, os.O_RDONLY, 0)
				if err != nil {
					return err
				}
				session.tmpErr, err = os.OpenFile(session.TmpErrName, os.O_RDONLY, 0)
				if err != nil {
					return err
				}

				s.sessions[session.Name] = session
				s.sessionsByPriority[session.PriorityOrder] = append(s.sessionsByPriority[session.PriorityOrder], session)
			} else {
				// update session
				exists.Started = session.Started
				exists.Ended = session.Ended
			}

		}

		priorities := collections.Keys(&s.sessionsByPriority)
		slices.Sort(priorities)
	end:
		for _, priority := range priorities {
			sessions, ok := s.sessionsByPriority[priority]
			if ok {
				for _, session := range sessions {
					fmt.Printf("electing session: [%s] ; ended: %v ; flushed: %v\n", session.Name, session.Ended, session.flushed)

					if !session.Started || session.Ended && session.flushed {
						continue
					}
					fmt.Printf("elected session: [%s]\n", session.Name)
					s.electedSession = session
					break end
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
	fmt.Printf("flushtailing session: [%s] ; ended: %v ; flushed: %v\n", s.electedSession.Name, s.electedSession.Ended, s.electedSession.flushed)

	buf := make([]byte, bufLen)
	n, err := filez.PartialCopy(s.electedSession.tmpOut, s.outputs.Out(), buf, s.electedSession.cursorOut, -1)
	if err != nil {
		return err
	}
	s.electedSession.cursorOut += int64(n)

	n, err = filez.PartialCopy(s.electedSession.tmpErr, s.outputs.Err(), buf, s.electedSession.cursorErr, -1)
	if err != nil {
		return err
	}
	s.electedSession.cursorErr += int64(n)

	if s.electedSession.Ended {
		s.electedSession.flushed = true
		s.electedSession = nil
		// FIXME: do not use a recusrsive call
		s.Flush()
	}

	return
}

func (*screenTailer) AsyncFlush(timeout time.Duration) (err error) {
	// Launch goroutine wich will continuously flush async display
	// TODO later ?
	return
}

func (*screenTailer) BlockTail(timeout time.Duration) (err error) {
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
		//sessionsByPriority: make(map[int][]*session),
	}
}

func NewAsyncScreenTailer(outputs printz.Outputs, tmpPath string) *screenTailer {
	if ok, _ := filez.IsDirectory(tmpPath); !ok {
		panic(fmt.Sprintf("unable to create read only async screen: [%s] path do not exists", tmpPath))
	}

	return &screenTailer{
		outputs:            outputs,
		tmpPath:            tmpPath,
		sessions:           make(map[string]*session),
		sessionsByPriority: make(map[int][]*session),
	}
}
