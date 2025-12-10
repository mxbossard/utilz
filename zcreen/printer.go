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
	printer0

	closeFilepath string
}

func (p *fsPrinter) Close(message string) error {
	err := filez.WriteString(p.closeFilepath, message, filez.DefaultFilePerms)
	if err != nil {
		return err
	}
	err = p.printer0.Close(message)
	return err
}

func buildFsOutputs(outFilepath, errFilepath string) (printz.Outputs, *os.File, *os.File) {
	filez.TouchMkdirAllOrPanic(outFilepath)
	filez.TouchMkdirAllOrPanic(errFilepath)
	outFile := filez.OpenOrPanic(outFilepath, os.O_WRONLY+os.O_APPEND, filez.DefaultFilePerms)
	errFile := filez.OpenOrPanic(errFilepath, os.O_WRONLY+os.O_APPEND, filez.DefaultFilePerms)
	outputs := printz.NewOutputs(outFile, errFile)
	return outputs, outFile, errFile
}

func buildFsPrinter(name string, priority int, outFilepath, errFilepath, closeFilepath string, opened bool) *fsPrinter {
	outputs, outFile, errFile := buildFsOutputs(outFilepath, errFilepath)
	prtr := printz.New(outputs)
	closingPrtr := printz.Closing(prtr)
	p := printer0{
		ClosingPrinter: closingPrtr,
		name:           name,
		tmpOut:         outFile,
		tmpErr:         errFile,
		// open:           opened,
		priorityOrder: priority,
	}
	return &fsPrinter{printer0: p, closeFilepath: closeFilepath}
}

func buildNotifierPrinter(zcreenPath, sessionName string, sessionPriority int, opened bool) *fsPrinter {
	notifierOutFilepath, notifierErrFilepath, closeFilepath := forgeNotiferPrinterFilepathes(zcreenPath, sessionName, sessionPriority)
	return buildFsPrinter("__notifier", 0, notifierOutFilepath, notifierErrFilepath, closeFilepath, opened)
}

func buildSessionPrinter(zcreenPath, sessionName string, sessionPriority int, printerName string, printerPriority int, opened bool) *fsPrinter {
	printerOutFilepath, printerErrFilepath, closeFilepath := forgeSessionPrinterFilepathes(zcreenPath, sessionName, sessionPriority, printerName, printerPriority)
	return buildFsPrinter(printerName, printerPriority, printerOutFilepath, printerErrFilepath, closeFilepath, opened)
}
