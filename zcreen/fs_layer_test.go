package zcreen

import (
	"os"
	"strings"
	"testing"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/printz"
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

	var err error

	globalNotifier := buildNotifierPrinter(tmpDir, "", 0, true)
	globalNotifier.Out("msg_globalNotifier\n")
	err = globalNotifier.Flush()
	assert.NoError(t, err)
	alpha1Foo1Printer1 := buildSessionPrinter(tmpDir, "alpha", 1, "foo", 1, true)
	alpha1Foo1Printer1.Out("msg_alpha1Foo1Printer1\n")
	err = alpha1Foo1Printer1.Flush()
	assert.NoError(t, err)
	alpha1Bar1Printer1 := buildSessionPrinter(tmpDir, "alpha", 1, "bar", 1, false)
	alpha1Bar1Printer1.Out("msg_alpha1Bar1Printer1\n")
	err = alpha1Bar1Printer1.Flush()
	assert.NoError(t, err)
	alpha1Bar1Printer2 := buildSessionPrinter(tmpDir, "alpha", 1, "bar", 1, false)
	alpha1Bar1Printer2.Out("msg_alpha1Bar1Printer2\n")
	err = alpha1Bar1Printer2.Flush()
	assert.NoError(t, err)
	alpha1Bar1Printer3 := buildSessionPrinter(tmpDir, "alpha", 1, "bar", 1, false)
	alpha1Bar1Printer3.Out("msg_alpha1Bar1Printer3\n")
	err = alpha1Bar1Printer3.Flush()
	assert.NoError(t, err)
	alpha1Bar1Printer4 := buildSessionPrinter(tmpDir, "alpha", 1, "bar", 1, false)
	alpha1Bar1Printer4.Out("msg_alpha1Bar1Printer4\n")
	err = alpha1Bar1Printer4.Flush()
	assert.NoError(t, err)
	alpha1Baz2Printer1 := buildSessionPrinter(tmpDir, "alpha", 1, "baz", 2, false)
	alpha1Baz2Printer1.Out("msg_alpha1Baz2Printer1\n")
	err = alpha1Baz2Printer1.Flush()
	assert.NoError(t, err)
	bravo1Notifier1 := buildNotifierPrinter(tmpDir, "bravo", 1, false)
	bravo1Notifier1.Out("msg_bravo1Notifier1\n")
	err = bravo1Notifier1.Flush()
	assert.NoError(t, err)
	bravo1Notifier2 := buildNotifierPrinter(tmpDir, "bravo", 1, false)
	bravo1Notifier2.Out("msg_bravo1Notifier2\n")
	err = bravo1Notifier2.Flush()
	assert.NoError(t, err)
	bravo1Notifier3 := buildNotifierPrinter(tmpDir, "bravo", 1, false)
	bravo1Notifier3.Out("msg_bravo1Notifier3\n")
	err = bravo1Notifier3.Flush()
	assert.NoError(t, err)
	bravo1Foo1Printer1 := buildSessionPrinter(tmpDir, "bravo", 1, "foo", 1, false)
	bravo1Foo1Printer1.Out("msg_bravo1Foo1Printer1\n")
	err = bravo1Foo1Printer1.Flush()
	assert.NoError(t, err)
	bravo1Foo1Printer2 := buildSessionPrinter(tmpDir, "bravo", 1, "foo", 1, false)
	bravo1Foo1Printer2.Out("msg_bravo1Foo1Printer2\n")
	err = bravo1Foo1Printer2.Flush()
	assert.NoError(t, err)

	zg := buildZcreenGroup(tmpDir)
	err = zg.scanFiles()
	require.NoError(t, err)
	require.NotNil(t, zg.notifierParts)
	assert.Len(t, zg.notifierParts.partsByKey, 1)
	require.NotNil(t, zg.sessionsByPrioName)
	assert.Len(t, zg.sessionsByPrioName, 2)

	alphaSg := zg.sessionsByPrioName[forgePrioNameKey(1, "alpha")]
	require.NotNil(t, alphaSg)
	require.NotNil(t, alphaSg.notifierParts)
	assert.Len(t, alphaSg.notifierParts.partsByKey, 0)
	require.NotNil(t, alphaSg.printersByPrioName)
	assert.Len(t, alphaSg.printersByPrioName, 3)

	alpha_1_foo_pg := alphaSg.printersByPrioName[forgePrioNameKey(1, "foo")]
	require.NotNil(t, alpha_1_foo_pg)
	assert.Len(t, alpha_1_foo_pg.partsByKey, 1)

	alpha_1_bar_pg := alphaSg.printersByPrioName[forgePrioNameKey(1, "bar")]
	require.NotNil(t, alpha_1_bar_pg)
	assert.Len(t, alpha_1_bar_pg.partsByKey, 4)

	alpha_2_baz_pg := alphaSg.printersByPrioName[forgePrioNameKey(2, "baz")]
	require.NotNil(t, alpha_2_baz_pg)
	assert.Len(t, alpha_2_baz_pg.partsByKey, 1)

	bravoSg := zg.sessionsByPrioName[forgePrioNameKey(1, "bravo")]
	require.NotNil(t, bravoSg)
	require.NotNil(t, bravoSg.notifierParts)
	assert.Len(t, bravoSg.notifierParts.partsByKey, 3)
	require.NotNil(t, bravoSg.printersByPrioName)
	assert.Len(t, bravoSg.printersByPrioName, 1)
	bravo_1_foo_pg := bravoSg.printersByPrioName[forgePrioNameKey(1, "foo")]
	require.NotNil(t, bravo_1_foo_pg)
	assert.Len(t, bravo_1_foo_pg.partsByKey, 2)

	zgo := zg.Outputer()
	require.NotNil(t, zgo)
	assert.Len(t, zgo.Sessions(), 2)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)

	out2W := &strings.Builder{}
	err2W := &strings.Builder{}
	outs2 := printz.NewOutputs(out2W, err2W)

	// alpha session outputs
	outW.Reset()
	errW.Reset()
	alphaOutputer := zgo.SessionOutputer("alpha")
	require.NotNil(t, alphaOutputer)
	assert.True(t, alphaOutputer.HasNext())
	assert.True(t, alphaOutputer.HasNext())
	alphaOutputer.Outputs(outs, "alpha")
	assert.True(t, alphaOutputer.HasNext())
	assert.Equal(t, "msg_globalNotifier\nmsg_alpha1Bar1Printer1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Re outputing should not print more data
	alphaOutputer.Outputs(outs, "alpha")
	assert.True(t, alphaOutputer.HasNext())
	assert.Equal(t, "msg_globalNotifier\nmsg_alpha1Bar1Printer1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Use a second outputer should print all available session datas without global notifications
	outW.Reset()
	errW.Reset()
	alphaOutputer2 := zgo.SessionOutputer("alpha")
	require.NotNil(t, alphaOutputer2)
	assert.True(t, alphaOutputer2.HasNext())
	assert.True(t, alphaOutputer2.HasNext())
	alphaOutputer2.Outputs(outs2, "alpha")
	assert.True(t, alphaOutputer2.HasNext())
	assert.Equal(t, "msg_alpha1Bar1Printer1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// bravo session outputs
	outW.Reset()
	errW.Reset()
	bravoOutputer := zgo.SessionOutputer("bravo")
	require.NotNil(t, bravoOutputer)
	assert.True(t, bravoOutputer.HasNext())
	assert.True(t, bravoOutputer.HasNext())
	bravoOutputer.Outputs(outs, "bravo")
	assert.True(t, bravoOutputer.HasNext())
	assert.Equal(t, "msg_bravo1Notifier1\nmsg_bravo1Notifier2\nmsg_bravo1Notifier3\nmsg_bravo1Foo1Printer1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Re outputing should not print more data
	bravoOutputer.Outputs(outs, "bravo")
	assert.True(t, bravoOutputer.HasNext())
	assert.Equal(t, "msg_bravo1Notifier1\nmsg_bravo1Notifier2\nmsg_bravo1Notifier3\nmsg_bravo1Foo1Printer1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Closing printer should print more data
	err = alpha1Foo1Printer1.Close("close1")
	assert.NoError(t, err)
	alphaOutputer.Outputs(outs, "alpha")
	assert.True(t, alphaOutputer.HasNext())
	assert.Equal(t, "msg_globalNotifier\nmsg_alpha1Bar1Printer1\nmsg_alpha1Bar1Printer1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Closing last printer part should print it first
	err = alpha1Bar1Printer4.Close("close2")
	assert.NoError(t, err)
	err = alpha1Baz2Printer1.Close("close3")
	assert.NoError(t, err)
	alphaOutputer.Outputs(outs, "alpha")
	assert.True(t, alphaOutputer.HasNext())
	assert.Equal(t, "msg_globalNotifier\nmsg_alpha1Bar1Printer1\nmsg_alpha1Bar1Printer4\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Closing all printer parts should print next printer
	err = alpha1Bar1Printer3.Close("close4")
	assert.NoError(t, err)
	err = alpha1Bar1Printer2.Close("close5")
	assert.NoError(t, err)
	alphaOutputer.Outputs(outs, "alpha")
	assert.True(t, alphaOutputer.HasNext())
	assert.Equal(t, "msg_globalNotifier\nmsg_alpha1Bar1Printer1\nmsg_alpha1Bar1Printer4\nmsg_alpha1Bar1Printer2\nmsg_alpha1Bar1Printer3\nmsg_alpha1Baz2Printer1\n", outW.String())
	assert.Equal(t, "", errW.String())
}
