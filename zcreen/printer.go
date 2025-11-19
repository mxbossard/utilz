package zcreen

import (
	"os"

	"github.com/mxbossard/utilz/printz"
)

type printer struct {
	printz.ClosingPrinter

	name string
	// open bool
	open          bool
	closeMessage  string
	consolidated  bool
	priorityOrder int

	tmpOut, tmpErr       *os.File
	cursorOut, cursorErr int64
}

func (p *printer) Close(message string) error {
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
