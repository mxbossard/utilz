package zcreen

import (
	"os"
	"testing"

	"github.com/mxbossard/utilz/filez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFsLayer_ScanNotExistingDir(t *testing.T) {
	zg := buildZcreenGroup("doNotExistsDir")
	err := zg.scanFiles()
	assert.NoError(t, err)
}

func TestFsLayer_EmptyDir(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestFsLayer_EmptyDir"
	require.NoError(t, os.RemoveAll(tmpDir))
	filez.MkdirAll(tmpDir, filez.DefaultDirPerms)
	defer os.RemoveAll(tmpDir)

	zg := buildZcreenGroup(tmpDir)
	err := zg.scanFiles()
	assert.NoError(t, err)
}

func TestFsLayer_AllFiles(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestFsLayer_AllFiles"
	require.NoError(t, os.RemoveAll(tmpDir))
	filez.MkdirAll(tmpDir, filez.DefaultDirPerms)
	//defer os.RemoveAll(tmpDir)

	globalNotifier := buildNotifierPrinter(tmpDir, "", 0, false)
	globalNotifier.Out("msg")
	globalNotifier.Flush()
	alpha1Foo1Printer1 := buildSessionPrinter(tmpDir, "alpha", 1, "foo", 1, false)
	alpha1Foo1Printer1.Out("msg")
	alpha1Foo1Printer1.Flush()
	alpha1Bar1Printer1 := buildSessionPrinter(tmpDir, "alpha", 1, "bar", 1, false)
	alpha1Bar1Printer1.Out("msg")
	alpha1Bar1Printer1.Flush()
	alpha1Bar1Printer2 := buildSessionPrinter(tmpDir, "alpha", 1, "bar", 1, false)
	alpha1Bar1Printer2.Out("msg")
	alpha1Bar1Printer2.Flush()
	alpha1Baz2Printer1 := buildSessionPrinter(tmpDir, "alpha", 1, "baz", 2, false)
	alpha1Baz2Printer1.Out("msg")
	alpha1Baz2Printer1.Flush()
	bravo1Notifier1 := buildNotifierPrinter(tmpDir, "bravo", 1, false)
	bravo1Notifier1.Out("msg")
	bravo1Notifier1.Flush()
	bravo1Notifier2 := buildNotifierPrinter(tmpDir, "bravo", 1, false)
	bravo1Notifier2.Out("msg")
	bravo1Notifier2.Flush()
	bravo1Notifier3 := buildNotifierPrinter(tmpDir, "bravo", 1, false)
	bravo1Notifier3.Out("msg")
	bravo1Notifier3.Flush()
	bravo1Foo1Printer1 := buildSessionPrinter(tmpDir, "bravo", 1, "foo", 1, false)
	bravo1Foo1Printer1.Out("msg")
	bravo1Foo1Printer1.Flush()
	bravo1Foo1Printer2 := buildSessionPrinter(tmpDir, "bravo", 1, "foo", 1, false)
	bravo1Foo1Printer2.Out("msg")
	bravo1Foo1Printer2.Flush()

	zg := buildZcreenGroup(tmpDir)
	err := zg.scanFiles()
	require.NoError(t, err)
	assert.Len(t, zg.notifierParts.partsByKey, 1)
	assert.Len(t, zg.sessionsByPrioName, 1)
	assert.Len(t, zg.sessionsByPrioName[1], 2)

	alphaSg := zg.sessionsByPrioName[1]["alpha"]
	require.NotNil(t, alphaSg)
	assert.Len(t, alphaSg.notifierParts.partsByKey, 0)
	assert.Len(t, alphaSg.printersByPrioName, 2)
	assert.Len(t, alphaSg.printersByPrioName[1], 2)
	assert.Len(t, alphaSg.printersByPrioName[1]["foo"], 1)
	assert.Len(t, alphaSg.printersByPrioName[1]["bar"], 2)
	assert.Len(t, alphaSg.printersByPrioName[2], 1)
	assert.Len(t, alphaSg.printersByPrioName[2]["baz"], 1)

	bravoSg := zg.sessionsByPrioName[1]["bravo"]
	require.NotNil(t, bravoSg)
	assert.Len(t, bravoSg.notifierParts.partsByKey, 3)
	assert.Len(t, bravoSg.printersByPrioName, 1)
	assert.Len(t, bravoSg.printersByPrioName[1], 1)
	assert.Len(t, bravoSg.printersByPrioName[1]["foo"], 2)
}
