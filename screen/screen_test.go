package screen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mby.fr/utils/filez"
	"mby.fr/utils/printz"
)

func TestGetScreen(t *testing.T) {
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	s := NewScreen(outs)
	assert.NotNil(t, s)
	assert.Implements(t, (*Sink)(nil), s)
}

func TestGetTailer(t *testing.T) {
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo40a"
	require.NoError(t, os.RemoveAll(tmpDir))
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err)
	s := NewAsyncScreenTailer(outs, tmpDir)
	assert.NotNil(t, s)
	assert.Implements(t, (*Tailer)(nil), s)
}

func TestGetAsyncScreen(t *testing.T) {
	tmpDir := "/tmp/foo40b"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	assert.NotNil(t, s)
	assert.DirExists(t, tmpDir)

	require.Panics(t, func() {
		NewAsyncScreen(tmpDir)
	})
}

func TestGetReadOnlyAsyncScreen(t *testing.T) {
	tmpDir := "/tmp/foo40c"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	assert.NotNil(t, s)
	assert.DirExists(t, tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	s2 := NewAsyncScreenTailer(outs, tmpDir)
	assert.NotNil(t, s2)

	require.Panics(t, func() {
		NewAsyncScreenTailer(outs, "/tmp/notExistingDir")
	})

}

func TestScreenGetSession(t *testing.T) {
	tmpDir := "/tmp/foo1001"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	require.NotNil(t, s)
	session := s.Session("bar1001", 42)
	assert.NotNil(t, session)
	assert.DirExists(t, tmpDir)
	assert.NoDirExists(t, tmpDir+"/"+sessionDirPrefix+"bar1001")

	err := session.Start(100 * time.Millisecond)
	require.NoError(t, err)
	assert.DirExists(t, tmpDir+"/"+sessionDirPrefix+"bar1001")
	matches, err := filepath.Glob(tmpDir + "/bar1001" + outFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	matches, err = filepath.Glob(tmpDir + "/bar1001" + errFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestScreenGetPrinter(t *testing.T) {
	tmpDir := "/tmp/foo2001"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	require.NotNil(t, s)
	session := s.Session("foo2001", 42)
	session.Start(100 * time.Millisecond)
	require.NotNil(t, session)
	prtr := session.Printer("bar", 10)
	assert.NotNil(t, prtr)
}

func TestAsyncScreen_Basic(t *testing.T) {
	tmpDir := "/tmp/foo3001"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)
	require.NotNil(t, screen)

	expectedSession := "foo3001"
	expectedPrinter := "bar"
	expectedMessage := "baz"

	session := screen.Session(expectedSession, 42)
	require.NotNil(t, session)
	err := session.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr10 := session.Printer(expectedPrinter, 10)
	require.NotNil(t, prtr10)

	sessionSerFilepath := filepath.Join(tmpDir, expectedSession+serializedExtension)
	sessionDirTmpFilepath := filepath.Join(tmpDir, sessionDirPrefix+expectedSession)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(tmpDir + "/" + expectedSession + outFileNameSuffix)
		return matches[0]
	}()
	sessionTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(tmpDir + "/" + expectedSession + errFileNameSuffix)
		return matches[0]
	}()
	printerTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedPrinter + outFileNameSuffix)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedPrinter + errFileNameSuffix)
		return matches[0]
	}()

	require.DirExists(t, tmpDir)
	require.DirExists(t, sessionDirTmpFilepath)
	assert.FileExists(t, sessionSerFilepath)
	assert.FileExists(t, sessionTmpOutFilepath)
	assert.FileExists(t, sessionTmpErrFilepath)
	assert.FileExists(t, printerTmpOutFilepath)
	assert.FileExists(t, printerTmpErrFilepath)

	assert.NotEmpty(t, func() string { s, _ := filez.ReadString(sessionSerFilepath); return s })
	ser, err := deserializeSession(sessionSerFilepath)
	require.NoError(t, err)
	assert.NotNil(t, ser)

	prtr10.Out(expectedMessage)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = prtr10.Flush()
	assert.NoError(t, err)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.NotEmpty(t, outW.String())
	assert.Equal(t, expectedMessage, outW.String())
	assert.Empty(t, errW.String())
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

}

func TestAsyncScreen_MultiplePrinters(t *testing.T) {
	tmpDir := "/tmp/foo4001"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSession := "foo4001"
	expectedPrinter10a := "bar10a"
	expectedPrinter20a := "bar20a"
	expectedPrinter20b := "bar20b"
	expectedPrinter30a := "bar30a"

	session := screen.Session(expectedSession, 42)
	err := session.Start(100 * time.Millisecond)
	assert.NoError(t, err)

	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(tmpDir + "/" + expectedSession + outFileNameSuffix)
		return matches[0]
	}()
	assert.FileExists(t, sessionTmpOutFilepath)

	prtr10a := session.Printer(expectedPrinter10a, 10)
	prtr20a := session.Printer(expectedPrinter20a, 20)
	prtr20b := session.Printer(expectedPrinter20b, 20)
	prtr30a := session.Printer(expectedPrinter30a, 30)

	prtr20a.Out("20a-1,")
	prtr20a.Out("20a-2,")
	prtr10a.Out("10a-1,")
	prtr30a.Out("30a-1,")
	prtr10a.Out("10a-2,")
	prtr20b.Out("20b-1,")
	prtr20a.Out("20a-3,")

	// First flush, nothing is Closed => first printers should be written only
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	err = screenTailer.Flush()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,", outW.String())
	assert.Empty(t, errW.String())

	// Close a printer which is not first => nothing more should be written
	session.ClosePrinter(expectedPrinter20a)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.Flush()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,", outW.String())
	assert.Empty(t, errW.String())

	prtr10a.Out("10a-3,")
	prtr20b.Out("20b-2,")

	// Close first printer => should write next printers
	session.ClosePrinter(expectedPrinter10a)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.Flush()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", outW.String())
	assert.Empty(t, errW.String())

	// Close a printer which is not first => should not write anything
	session.ClosePrinter(expectedPrinter30a)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.Flush()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", outW.String())
	assert.Empty(t, errW.String())

	// Close last printer => should write everything
	session.ClosePrinter(expectedPrinter20b)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,"+"30a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.Flush()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,"+"30a-1,", outW.String())
	assert.Empty(t, errW.String())
}

func TestAsyncScreen_MultipleSessions(t *testing.T) {
	tmpDir := "/tmp/foo5001c"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSessionA := "barA5001"
	expectedPrinterA10a := "barA10a"
	expectedPrinterA20a := "barA20a"
	expectedSessionB := "barB5001"
	expectedPrinterB10a := "barB10a"
	expectedPrinterB30a := "barB30a"
	expectedSessionC := "barC5001"
	expectedPrinterC10a := "barC10a"
	expectedPrinterC30a := "barC30a"
	expectedPrinterC40a := "barC40a"

	sessionA := screen.Session(expectedSessionA, 12)
	err := sessionA.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrA10a := sessionA.Printer(expectedPrinterA10a, 10)
	prtrA20a := sessionA.Printer(expectedPrinterA20a, 20)

	sessionB := screen.Session(expectedSessionB, 42)
	err = sessionB.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrB10a := sessionB.Printer(expectedPrinterB10a, 10)
	prtrB30a := sessionB.Printer(expectedPrinterB30a, 30)

	sessionC := screen.Session(expectedSessionC, 42)
	err = sessionC.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrC10a := sessionC.Printer(expectedPrinterC10a, 10)
	prtrC30a := sessionC.Printer(expectedPrinterC30a, 30)
	prtrC40a := sessionC.Printer(expectedPrinterC40a, 40)

	assert.Empty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))
	assert.Empty(t, filez.ReadStringOrPanic(sessionB.TmpOutName))

	prtrB10a.Out("B10a1,")
	prtrB30a.Out("B30a1,")
	prtrA10a.Out("A10a1,")
	prtrB30a.Out("B30a2,")
	err = sessionB.ClosePrinter(expectedPrinterB30a)
	assert.NoError(t, err)
	prtrB10a.Out("B10a2,")
	err = sessionB.ClosePrinter(expectedPrinterB10a)
	assert.NoError(t, err)
	err = sessionB.End()
	assert.NoError(t, err)
	err = sessionB.End()
	assert.NoError(t, err)
	prtrA20a.Out("A20a1,")
	prtrA20a.Out("A20a2,")
	err = sessionA.ClosePrinter(expectedPrinterA20a)
	assert.NoError(t, err)
	prtrA10a.Out("A10a2,")
	err = sessionA.ClosePrinter(expectedPrinterA10a)
	assert.NoError(t, err)

	prtrC10a.Out("C10a1,")
	prtrC30a.Out("C30a1,")
	prtrC40a.Out("C40a1,")

	assert.Empty(t, outW.String())
	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())

	assert.Empty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))
	assert.NotEmpty(t, filez.ReadStringOrPanic(sessionB.TmpOutName))

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())

	err = sessionA.Flush()
	assert.NoError(t, err)

	assert.NotEmpty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,", outW.String())

	err = sessionA.End()
	assert.NoError(t, err)

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,B10a1,B10a2,B30a1,B30a2,", outW.String())

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,B10a1,B10a2,B30a1,B30a2,", outW.String())

	assert.Equal(t, "", filez.ReadStringOrPanic(sessionC.TmpOutName))
	err = sessionC.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "C10a1,", filez.ReadStringOrPanic(sessionC.TmpOutName))
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,"+"B10a1,B10a2,B30a1,B30a2,", outW.String())

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,", outW.String())

	err = sessionC.ClosePrinter(expectedPrinterC10a)
	assert.NoError(t, err)
	err = sessionC.Flush()
	assert.NoError(t, err)

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,C30a1,", outW.String())

	err = sessionC.End()
	assert.NoError(t, err)
	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,C30a1,C40a1,", outW.String())
}

func TestAsyncScreen_Notifications(t *testing.T) {
	tmpDir := "/tmp/foo6001"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSessionA := "barA6001"
	expectedPrinterA10a := "barA10a"
	expectedPrinterA20a := "barA20a"
	expectedSessionB := "barB6001"
	expectedPrinterB10a := "barB10a"
	expectedPrinterB30a := "barB30a"
	expectedSessionC := "barC6001"
	expectedPrinterC10a := "barC20a"
	expectedPrinterC30a := "barC30a"
	expectedPrinterC40a := "barC40a"

	sessionA := screen.Session(expectedSessionA, 12)
	err := sessionA.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrA10a := sessionA.Printer(expectedPrinterA10a, 10)
	prtrA20a := sessionA.Printer(expectedPrinterA20a, 20)

	sessionB := screen.Session(expectedSessionB, 42)
	err = sessionB.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrB10a := sessionB.Printer(expectedPrinterB10a, 10)
	prtrB30a := sessionB.Printer(expectedPrinterB30a, 30)

	sessionC := screen.Session(expectedSessionC, 42)
	err = sessionC.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrC10a := sessionC.Printer(expectedPrinterC10a, 10)
	prtrC30a := sessionC.Printer(expectedPrinterC30a, 30)
	prtrC40a := sessionC.Printer(expectedPrinterC40a, 40)

	assert.Empty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))
	assert.Empty(t, filez.ReadStringOrPanic(sessionB.TmpOutName))

	screen.NotifyPrinter().Out("notif1,")

	prtrB10a.Out("B10a1,")
	prtrB30a.Out("B30a1,")
	prtrA10a.Out("A10a1,")
	prtrB30a.Out("B30a2,")
	err = sessionB.ClosePrinter(expectedPrinterB30a)
	assert.NoError(t, err)
	prtrB10a.Out("B10a2,")
	err = sessionB.ClosePrinter(expectedPrinterB10a)
	assert.NoError(t, err)
	err = sessionB.End()
	assert.NoError(t, err)
	prtrA20a.Out("A20a1,")
	prtrA20a.Out("A20a2,")
	err = sessionA.ClosePrinter(expectedPrinterA20a)
	assert.NoError(t, err)
	prtrA10a.Out("A10a2,")
	err = sessionA.ClosePrinter(expectedPrinterA10a)
	assert.NoError(t, err)

	screen.NotifyPrinter().Out("notif2,")
	err = screen.NotifyPrinter().Flush()
	assert.NoError(t, err)
	screen.NotifyPrinter().Out("notif3,")

	prtrC10a.Out("C10a1,")
	prtrC30a.Out("C30a1,")
	prtrC40a.Out("C40a1,")

	assert.Empty(t, outW.String())
	err = screenTailer.Flush()
	assert.NoError(t, err)
	screen.NotifyPrinter().Out("notif4,")
	err = screen.NotifyPrinter().Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,", outW.String())

	assert.Empty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))
	assert.NotEmpty(t, filez.ReadStringOrPanic(sessionB.TmpOutName))

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,", outW.String())

	err = sessionA.Flush()
	assert.NoError(t, err)

	assert.NotEmpty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,", outW.String())

	err = sessionA.End()
	assert.NoError(t, err)

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,", outW.String())

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,", outW.String())

	assert.Equal(t, "", filez.ReadStringOrPanic(sessionC.TmpOutName))
	err = sessionC.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "C10a1,", filez.ReadStringOrPanic(sessionC.TmpOutName))
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,", outW.String())

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,", outW.String())

	screen.NotifyPrinter().Out("notif5,")
	err = screen.NotifyPrinter().Flush()
	assert.NoError(t, err)

	err = sessionC.ClosePrinter(expectedPrinterC10a)
	assert.NoError(t, err)
	err = sessionC.Flush()
	assert.NoError(t, err)

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,C30a1,", outW.String())

	err = sessionC.End()
	assert.NoError(t, err)
	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,C30a1,C40a1,"+"notif5,", outW.String())
}
