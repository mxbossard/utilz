package zcreen

import (
	"os"

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

	outFilepath   string
	errFilepath   string
	closeFilepath string
}

func (p *fsPrinter) Close(message string) (err error) {
	nOut, nErr := p.ClosingPrinter.Counts()
	if nOut > 0 || nErr > 0 {
		// Close printer only if flushed
		err := filez.WriteString(p.closeFilepath, message, filez.DefaultFilePerms)
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

func buildFsPrinter(name string, priority int, outFilepath, errFilepath, closeFilepath string, opened bool) *fsPrinter {
	outputs, _, _ := buildFsOutputs(outFilepath, errFilepath)
	prtr := printz.New(outputs)
	closingPrtr := printz.Closing(prtr)
	return &fsPrinter{
		ClosingPrinter: closingPrtr,
		name:           name,
		priorityOrder:  priority,
		outFilepath:    outFilepath,
		errFilepath:    errFilepath,
		closeFilepath:  closeFilepath,
	}
}

func buildNotifierPrinter(zcreenPath, sessionName string, sessionPriority int, opened bool) *fsPrinter {
	notifierOutFilepath, notifierErrFilepath, closeFilepath := forgeNotiferPrinterFilepathes(zcreenPath, sessionName, sessionPriority)
	return buildFsPrinter("__notifier", 0, notifierOutFilepath, notifierErrFilepath, closeFilepath, opened)
}

func buildSessionPrinter(zcreenPath, sessionName string, sessionPriority int, printerName string, printerPriority int, opened bool) *fsPrinter {
	printerOutFilepath, printerErrFilepath, closeFilepath := forgeSessionPrinterFilepathes(zcreenPath, sessionName, sessionPriority, printerName, printerPriority)
	return buildFsPrinter(printerName, printerPriority, printerOutFilepath, printerErrFilepath, closeFilepath, opened)
}
