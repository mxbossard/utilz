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
	//outputs            printz.Outputs
	//openedSession      *session
	sessionsByPriority map[int][]*session
}

func (s *screen) Session(name string, priorityOrder int) *session {
	if session, ok := s.sessions[name]; ok {
		return session
	}

	sessionDirpath := filepath.Join(s.tmpPath, printerDirPrefix+name)
	if _, err := os.Stat(sessionDirpath); err == nil {
		panic(fmt.Sprintf("unable to create async screen session: [%s] path already exists", sessionDirpath))
	}

	_, tmpOut, tmpErr := buildTmpOutputs(s.tmpPath, name)
	session := &session{
		name:          name,
		priorityOrder: priorityOrder,
		readOnly:      false,
		tmpPath:       sessionDirpath,
		//suiteTmpOutputs: suiteTmpOutputs,
		tmpOutName: tmpOut.Name(),
		tmpErrName: tmpErr.Name(),
		tmpOut:     tmpOut,
		tmpErr:     tmpErr,
		// tmpOutputs:         make(map[string]printz.Outputs),
		// openedPrinters:     make(map[string]printz.Printer),
		printersByPriority: make(map[int][]*printer),
		printers:           make(map[string]*printer),
	}
	s.sessions[name] = session
	s.sessionsByPriority[priorityOrder] = append(s.sessionsByPriority[priorityOrder], session)

	return session
}

func (*screen) ConfigPrinter(name string) (s *session) {
	// TODO later
	return
}

type screenTailer struct {
	tmpPath            string
	outputs            printz.Outputs
	openedSession      *session
	sessions           map[string]*session
	sessionsByPriority map[int][]*session
}

func (s *screenTailer) Flush() (err error) {
	// Flush the display : print sessions in order onto std outputs (keep written bytes count)

	if s.openedSession == nil {
		// TODO: scan tmp dir to refresh maps sessions & sessionsByPriority
		// TODO: must build all sessions from files

		//dirs, err := os.ReadDir(s.tmpPath)
		sers, err := filepath.Glob(s.tmpPath + "/*.ser")
		if err != nil {
			return err
		}
		for _, filePath := range sers {
			// filePath := filepath.Join(s.tmpPath, filename)
			session, err := deserializeSession(filePath)
			if err != nil {
				return err
			}

			// clear session
			//session.printers = nil
			//session.printersByPriority = nil
			session.currentPriority = nil
			session.tmpOut = nil
			session.tmpErr = nil

			if exists, ok := s.sessions[session.name]; !ok {
				// init session
				outPath := filepath.Join(s.tmpPath, session.tmpOutName)
				errPath := filepath.Join(s.tmpPath, session.tmpErrName)
				session.tmpOut, err = os.OpenFile(outPath, os.O_RDONLY, 0)
				if err != nil {
					return err
				}
				session.tmpErr, err = os.OpenFile(errPath, os.O_RDONLY, 0)
				if err != nil {
					return err
				}

				s.sessions[session.name] = session
				s.sessionsByPriority[session.priorityOrder] = append(s.sessionsByPriority[session.priorityOrder], session)
			} else {
				// update session
				exists.started = session.started
				exists.ended = session.ended
			}

		}

		priorities := collections.Keys(&s.sessionsByPriority)
		slices.Sort(priorities)
		for _, priority := range priorities {
			sessions, ok := s.sessionsByPriority[priority]
			if ok {
				for _, session := range sessions {
					if !session.started || session.ended {
						continue
					}
					s.openedSession = session
				}
			}
		}
	}

	if s.openedSession == nil {
		return
	}

	buf := make([]byte, bufLen)
	n, err := filez.PartialCopy(s.openedSession.tmpOut, s.outputs.Out(), buf, s.openedSession.cursorOut, -1)
	if err != nil {
		return err
	}
	s.openedSession.cursorOut += int64(n)

	n, err = filez.PartialCopy(s.openedSession.tmpErr, s.outputs.Err(), buf, s.openedSession.cursorErr, -1)
	if err != nil {
		return err
	}
	s.openedSession.cursorErr += int64(n)

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
		tmpPath:            tmpPath,
		sessions:           make(map[string]*session),
		sessionsByPriority: make(map[int][]*session),
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
