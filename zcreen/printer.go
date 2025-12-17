package zcreen

import (
	"os"
	"time"

	"github.com/mxbossard/utilz/anzi"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/printz"
)

type printer0 struct {
	printz.ClosingPrinter

	name string
	// open          bool // FIXME: open is not used in printer => should be removed
	// closeMessage  string
	consolidated  bool
	priorityOrder int

	tmpOut, tmpErr       *os.File
	cursorOut, cursorErr int64
}

func (p *printer0) Close(message string) error {
	err := p.ClosingPrinter.Close(message)
	if err != nil {
		return err
	}

	if p.tmpOut != nil {
		err := p.tmpOut.Close()
		if err != nil {
			return err
		}
		// err = os.RemoveAll(p.tmpOut.Name())
		// if err != nil {
		// 	return err
		// }
	}

	if p.tmpErr != nil {
		err := p.tmpErr.Close()
		if err != nil {
			return err
		}
		// err = os.RemoveAll(p.tmpErr.Name())
		// if err != nil {
		// 	return err
		// }
	}

	return nil
}

type fsPrinter struct {
	printz.ClosingPrinter

	name          string
	priorityOrder int

	outFilepath    string
	errFilepath    string
	closedFilepath string
}

func (p *fsPrinter) Close(message string) (err error) {
	nOut, nErr := p.ClosingPrinter.Counts()
	if nOut > 0 || nErr > 0 {
		err = p.Flush()
		if err != nil {
			return err
		}
		// Close printer only if flushed
		err := filez.WriteString(p.closedFilepath, message, filez.DefaultFilePerms)
		if err != nil {
			return err
		}
		err = p.ClosingPrinter.Close(message)
	}
	return
}

func buildFsOutputs(outFilepath, errFilepath string) (printz.Outputs, *os.File, *os.File) {
	// filez.TouchMkdirAllOrPanic(outFilepath)
	// filez.TouchMkdirAllOrPanic(errFilepath)
	// outFile := filez.OpenOrPanic(outFilepath, os.O_WRONLY+os.O_APPEND, filez.DefaultFilePerms)
	// errFile := filez.OpenOrPanic(errFilepath, os.O_WRONLY+os.O_APPEND, filez.DefaultFilePerms)
	// outputs := printz.NewOutputs(outFile, errFile)
	outputs := printz.NewLazyFileOutputs(outFilepath, errFilepath, os.O_CREATE+os.O_WRONLY+os.O_APPEND, filez.DefaultFilePerms)
	return outputs, nil, nil
}

func buildFsPrinter(name string, priority int, outFilepath, errFilepath, closeFilepath string) *fsPrinter {
	outputs, _, _ := buildFsOutputs(outFilepath, errFilepath)
	prtr := printz.New(outputs)
	closingPrtr := printz.Closing(prtr)
	return &fsPrinter{
		ClosingPrinter: closingPrtr,
		name:           name,
		priorityOrder:  priority,
		outFilepath:    outFilepath,
		errFilepath:    errFilepath,
		closedFilepath: closeFilepath,
	}
}

// noPrintTimeout => auto close printer if nothing flushed before timeout.
// auto reopen new file on write if closed.
// if a file has a modTime older than timeout, can be considered closed.
type autoFsPrinter struct {
	*fsPrinter

	noPrintTimeout time.Duration
	filepathForge  func(qualifier string) string
}

func updateFileOutputs(p *autoFsPrinter) error {
	if p.ClosingPrinter != nil {
		lastFlush := p.fsPrinter.LastFlush()
		if !lastFlush.IsZero() && time.Since(p.fsPrinter.LastFlush()) > p.noPrintTimeout {
			// files output are outdated => close them and open new ones.
			err := p.fsPrinter.Close("outputs outdated")
			if err != nil {
				return err
			}
			p.ClosingPrinter = nil
		}
	}
	if p.ClosingPrinter == nil || p.IsClosed() {
		outFilepath := p.filepathForge(outFileQualifier)
		errFilepath := p.filepathForge(errFileQualifier)
		closedFilepath := p.filepathForge(closedFileQualifier)
		outputs, _, _ := buildFsOutputs(outFilepath, errFilepath)
		prtr := printz.New(outputs)
		closingPrtr := printz.Closing(prtr)
		p.fsPrinter.ClosingPrinter = closingPrtr
		p.fsPrinter.outFilepath = outFilepath
		p.fsPrinter.errFilepath = errFilepath
		p.fsPrinter.closedFilepath = closedFilepath
		p.Open()
	}
	return nil
}

func (p *autoFsPrinter) Close(msg string) error {
	if p.ClosingPrinter != nil {
		return p.fsPrinter.Close(msg)
	}
	return nil
}

func (p *autoFsPrinter) IsClosed() bool {
	if p.ClosingPrinter != nil {
		return p.fsPrinter.IsClosed()
	}
	return true
}

func (p *autoFsPrinter) Flush() error {
	// FIXME: Can I delegate output files creation on Flush instead of Print ?
	updateFileOutputs(p)
	return p.ClosingPrinter.Flush()
}

func (p *autoFsPrinter) Flushed() bool {
	if p.ClosingPrinter != nil {
		return p.fsPrinter.Flushed()
	}
	return true
}

func (p *autoFsPrinter) LastPrint() time.Time {
	if p.ClosingPrinter != nil {
		return p.fsPrinter.LastPrint()
	}
	return time.Time{}
}
func (p *autoFsPrinter) LastFlush() time.Time {
	if p.ClosingPrinter != nil {
		return p.fsPrinter.LastFlush()
	}
	return time.Time{}
}
func (p *autoFsPrinter) Counts() (int64, int64) {
	if p.ClosingPrinter != nil {
		return p.fsPrinter.Counts()
	}
	return 0, 0
}
func (p *autoFsPrinter) Outputs() printz.Outputs {
	updateFileOutputs(p)
	return p.fsPrinter.Outputs()
}

func (p *autoFsPrinter) RecoverableOut(obj ...interface{}) error {
	err := updateFileOutputs(p)
	if err != nil {
		return err
	}
	return p.fsPrinter.RecoverableOut(obj...)
}

func (p *autoFsPrinter) Out(obj ...interface{}) {
	err := updateFileOutputs(p)
	if err != nil {
		panic(err)
	}
	p.fsPrinter.Out(obj...)
}

func (p *autoFsPrinter) Outf(msg string, obj ...interface{}) {
	err := updateFileOutputs(p)
	if err != nil {
		panic(err)
	}
	p.fsPrinter.Outf(msg, obj...)
}

func (p *autoFsPrinter) ColoredOutf(color anzi.Color, msg string, obj ...interface{}) {
	err := updateFileOutputs(p)
	if err != nil {
		panic(err)
	}
	p.fsPrinter.ColoredOutf(color, msg, obj...)
}

func (p *autoFsPrinter) RecoverableErr(obj ...interface{}) error {
	err := updateFileOutputs(p)
	if err != nil {
		return err
	}
	return p.fsPrinter.RecoverableErr(obj...)
}

func (p *autoFsPrinter) Err(obj ...interface{}) {
	err := updateFileOutputs(p)
	if err != nil {
		panic(err)
	}
	p.fsPrinter.Err(obj...)
}

func (p *autoFsPrinter) Errf(msg string, obj ...interface{}) {
	err := updateFileOutputs(p)
	if err != nil {
		panic(err)
	}
	p.fsPrinter.Errf(msg, obj...)
}

func (p *autoFsPrinter) ColoredErrf(color anzi.Color, msg string, obj ...interface{}) {
	err := updateFileOutputs(p)
	if err != nil {
		panic(err)
	}
	p.fsPrinter.ColoredErrf(color, msg, obj...)
}

func buildAutoFsPrinter(name string, priority int, filepathForge func(qualifier string) string) *autoFsPrinter {
	fsPrinter := &fsPrinter{
		name:           name,
		priorityOrder:  priority,
		outFilepath:    filepathForge(outFileQualifier),
		errFilepath:    filepathForge(errFileQualifier),
		closedFilepath: filepathForge(closedFileQualifier),
	}

	return &autoFsPrinter{
		fsPrinter: fsPrinter,
		// noPrintTimeout: noPrintTimeout,
		filepathForge: filepathForge,
	}
}

func buildNotifierPrinter(zcreenPath, sessionName string, sessionPriority int) *fsPrinter {
	notifierOutFilepath, notifierErrFilepath, closeFilepath := forgeNotiferPrinterFilepathes(zcreenPath, sessionName, sessionPriority)
	return buildFsPrinter("__notifier", 0, notifierOutFilepath, notifierErrFilepath, closeFilepath)
}

func buildSessionPrinter(zcreenPath, sessionName string, sessionPriority int, printerName string, printerPriority int) *fsPrinter {
	printerOutFilepath, printerErrFilepath, closeFilepath := forgeSessionPrinterFilepathes(zcreenPath, sessionName, sessionPriority, printerName, printerPriority)
	return buildFsPrinter(printerName, printerPriority, printerOutFilepath, printerErrFilepath, closeFilepath)
}

func buildNotifierPrinter0(zcreenPath, sessionName string, sessionPriority int) *autoFsPrinter {
	notifierOutFilepath, notifierErrFilepath, notifierClosedFilepath := forgeNotiferPrinterFilepathes(zcreenPath, sessionName, sessionPriority)
	filepathForge := func(qualifier string) string {
		switch qualifier {
		case outFileQualifier:
			return notifierOutFilepath
		case errFileQualifier:
			return notifierErrFilepath
		case closedFileQualifier:
			return notifierClosedFilepath
		}
		return ""
	}
	return buildAutoFsPrinter("__notifier", 0, filepathForge)
}

func buildSessionPrinter0(zcreenPath, sessionName string, sessionPriority int, printerName string, printerPriority int) *autoFsPrinter {
	notifierOutFilepath, notifierErrFilepath, notifierClosedFilepath := forgeSessionPrinterFilepathes(zcreenPath, sessionName, sessionPriority, printerName, printerPriority)
	filepathForge := func(qualifier string) string {
		switch qualifier {
		case outFileQualifier:
			return notifierOutFilepath
		case errFileQualifier:
			return notifierErrFilepath
		case closedFileQualifier:
			return notifierClosedFilepath
		}
		return ""
	}
	return buildAutoFsPrinter(printerName, printerPriority, filepathForge)
}
