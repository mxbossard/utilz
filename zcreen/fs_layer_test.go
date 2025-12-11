package zcreen

import (
	"fmt"
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
	updated, err := zg.scanFiles()
	assert.False(t, updated)
	assert.NoError(t, err)
}

func TestFsLayer_EmptyDir(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestFsLayer_EmptyDir"
	require.NoError(t, os.RemoveAll(tmpDir))
	filez.MkdirAll(tmpDir, filez.DefaultDirPerms)
	defer os.RemoveAll(tmpDir)

	zg := buildZcreenGroup(tmpDir)
	updated, err := zg.scanFiles()
	assert.False(t, updated)
	assert.NoError(t, err)
}

func buildTestFsPrinter(t *testing.T, m *map[string]*fsPrinter, tmpDir, session string, sessionPrio int, name string, prio, id int) (key, msg string, p *fsPrinter) {
	if session == "" {
		key = fmt.Sprintf("global_notifier_%d", id)
		p = buildNotifierPrinter(tmpDir, "", 0, true)
	} else if name == "" {
		key = fmt.Sprintf("%s_%d_notifier_%d", session, sessionPrio, id)
		p = buildNotifierPrinter(tmpDir, session, sessionPrio, true)
	} else {
		key = fmt.Sprintf("%s_%d_%s_%d_%d", session, sessionPrio, name, prio, id)
		p = buildSessionPrinter(tmpDir, session, sessionPrio, name, prio, true)
	}
	(*m)[key] = p
	msg = fmt.Sprintf("msg-%s\n", key)
	p.Out(msg)
	err := p.Flush()
	assert.NoError(t, err)

	return
}

func buildTestFsPrinters_scenario1(t *testing.T, tmpDir string) (m map[string]*fsPrinter) {
	m = make(map[string]*fsPrinter)

	// Global notifiers
	buildTestFsPrinter(t, &m, tmpDir, "", 0, "", 0, 1)
	buildTestFsPrinter(t, &m, tmpDir, "", 0, "", 0, 2)
	buildTestFsPrinter(t, &m, tmpDir, "", 0, "", 0, 3)

	// alpha session
	// -- notifiers
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "", 0, 1)
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "", 0, 2)
	// -- printer foo
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "foo", 0, 1)
	// -- printer bar
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "bar", 0, 1)
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "bar", 0, 2)
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "bar", 0, 3)
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "bar", 0, 4)
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "bar", 0, 5)
	// -- printer baz
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "baz", 2, 1)
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "baz", 2, 2)
	buildTestFsPrinter(t, &m, tmpDir, "alpha", 5, "baz", 2, 3)

	// bravo session
	// -- notifiers
	buildTestFsPrinter(t, &m, tmpDir, "bravo", 0, "", 0, 1)
	buildTestFsPrinter(t, &m, tmpDir, "bravo", 0, "", 0, 2)
	buildTestFsPrinter(t, &m, tmpDir, "bravo", 0, "", 0, 3)
	// -- printer pif
	buildTestFsPrinter(t, &m, tmpDir, "bravo", 0, "pif", 0, 1)
	buildTestFsPrinter(t, &m, tmpDir, "bravo", 0, "pif", 0, 2)
	buildTestFsPrinter(t, &m, tmpDir, "bravo", 0, "pif", 0, 3)
	// -- printer paf
	buildTestFsPrinter(t, &m, tmpDir, "bravo", 0, "paf", 1, 1)

	// gamma session
	// -- notifiers
	//buildTestFsPrinter(t, &m, tmpDir, "gamma", 0, "", 0, 1)
	// -- printer bar
	buildTestFsPrinter(t, &m, tmpDir, "gamma", 0, "bar", 20, 1)
	// -- printer baz
	buildTestFsPrinter(t, &m, tmpDir, "gamma", 0, "baz", 10, 1)

	return
}

func TestFsLayer_ScanFiles(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestFsLayer_ScanFiles"
	require.NoError(t, os.RemoveAll(tmpDir))
	filez.MkdirAll(tmpDir, filez.DefaultDirPerms)
	defer os.RemoveAll(tmpDir)
	var err error

	scenario1 := buildTestFsPrinters_scenario1(t, tmpDir)
	_ = scenario1

	zg := buildZcreenGroup(tmpDir)
	updated, err := zg.scanFiles()
	assert.True(t, updated)
	require.NoError(t, err)
	require.NotNil(t, zg.notifierParts)

	// Test global notifiers
	assert.Len(t, zg.notifierParts.partsByKey, 3)
	require.NotNil(t, zg.sessionsByPrioName)
	assert.Len(t, zg.sessionsByPrioName, 3)

	// Test alpha session
	alphaSg := *zg.sessionsByPrioName[forgePrioNameKey(5, "alpha")]
	require.NotNil(t, alphaSg)

	require.NotNil(t, alphaSg.notifierParts)
	assert.Len(t, alphaSg.notifierParts.partsByKey, 2)

	require.NotNil(t, alphaSg.printersByPrioName)
	assert.Len(t, alphaSg.printersByPrioName, 3)

	alpha_foo_pg := alphaSg.printersByPrioName[forgePrioNameKey(0, "foo")]
	require.NotNil(t, alpha_foo_pg)
	assert.Len(t, alpha_foo_pg.partsByKey, 1)

	alpha_bar_pg := alphaSg.printersByPrioName[forgePrioNameKey(0, "bar")]
	require.NotNil(t, alpha_bar_pg)
	assert.Len(t, alpha_bar_pg.partsByKey, 5)

	alpha_baz_pg := alphaSg.printersByPrioName[forgePrioNameKey(2, "baz")]
	require.NotNil(t, alpha_baz_pg)
	assert.Len(t, alpha_baz_pg.partsByKey, 3)

	// Test bravo session
	bravoSg := zg.sessionsByPrioName[forgePrioNameKey(0, "bravo")]
	require.NotNil(t, bravoSg)

	require.NotNil(t, bravoSg.notifierParts)
	assert.Len(t, bravoSg.notifierParts.partsByKey, 3)

	require.NotNil(t, bravoSg.printersByPrioName)
	assert.Len(t, bravoSg.printersByPrioName, 2)

	bravo_pif_pg := bravoSg.printersByPrioName[forgePrioNameKey(0, "pif")]
	require.NotNil(t, bravo_pif_pg)
	assert.Len(t, bravo_pif_pg.partsByKey, 3)

	bravo_paf_pg := bravoSg.printersByPrioName[forgePrioNameKey(1, "paf")]
	require.NotNil(t, bravo_paf_pg)
	assert.Len(t, bravo_paf_pg.partsByKey, 1)

	// Test gamma session
	gammaSg := zg.sessionsByPrioName[forgePrioNameKey(0, "gamma")]
	require.NotNil(t, gammaSg)

	require.NotNil(t, gammaSg.notifierParts)
	assert.Len(t, gammaSg.notifierParts.partsByKey, 0)

	require.NotNil(t, gammaSg.printersByPrioName)
	assert.Len(t, gammaSg.printersByPrioName, 2)

	gamma_bar_pg := gammaSg.printersByPrioName[forgePrioNameKey(20, "bar")]
	require.NotNil(t, gamma_bar_pg)
	assert.Len(t, gamma_bar_pg.partsByKey, 1)

	gamma_baz_pg := gammaSg.printersByPrioName[forgePrioNameKey(10, "baz")]
	require.NotNil(t, gamma_baz_pg)
	assert.Len(t, gamma_baz_pg.partsByKey, 1)

}

func TestFsLayer_Sessions(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestFsLayer_Sessions"
	require.NoError(t, os.RemoveAll(tmpDir))
	filez.MkdirAll(tmpDir, filez.DefaultDirPerms)
	defer os.RemoveAll(tmpDir)
	var err error

	scenario1 := buildTestFsPrinters_scenario1(t, tmpDir)
	_ = scenario1

	zg := buildZcreenGroup(tmpDir)
	updated, err := zg.scanFiles()
	assert.True(t, updated)
	require.NoError(t, err)

	zgo := zg.outputer()
	require.NotNil(t, zgo)
	sessions := zgo.sessions()
	assert.Len(t, sessions, 3)
	assert.Equal(t, []string{"bravo", "gamma", "alpha"}, sessions)
}

func TestFsLayer_OuputsOrdering(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestFsLayer_OuputsOrdering"
	require.NoError(t, os.RemoveAll(tmpDir))
	filez.MkdirAll(tmpDir, filez.DefaultDirPerms)
	defer os.RemoveAll(tmpDir)
	var err error

	scenario1 := buildTestFsPrinters_scenario1(t, tmpDir)
	alphaBarPrinter1 := scenario1["alpha_5_bar_0_1"]
	alphaBarPrinter2 := scenario1["alpha_5_bar_0_2"]
	alphaBarPrinter3 := scenario1["alpha_5_bar_0_3"]
	alphaBarPrinter4 := scenario1["alpha_5_bar_0_4"]
	alphaBarPrinter5 := scenario1["alpha_5_bar_0_5"]
	alphaBazPrinter1 := scenario1["alpha_5_baz_2_1"]

	zg := buildZcreenGroup(tmpDir)
	updated, err := zg.scanFiles()
	assert.True(t, updated)
	require.NoError(t, err)

	zgo := zg.outputer()
	require.NotNil(t, zgo)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)

	out2W := &strings.Builder{}
	err2W := &strings.Builder{}
	outs2 := printz.NewOutputs(out2W, err2W)

	// alpha session outputs
	outW.Reset()
	errW.Reset()
	alphaOutputer := zgo.sessionOutputer("alpha")
	require.NotNil(t, alphaOutputer)
	assert.True(t, alphaOutputer.hasNext())
	assert.True(t, alphaOutputer.hasNext())
	err = alphaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, alphaOutputer.hasNext())
	assert.Equal(t, "msg-global_notifier_1\nmsg-global_notifier_2\nmsg-global_notifier_3\nmsg-alpha_5_notifier_1\nmsg-alpha_5_notifier_2\nmsg-alpha_5_bar_0_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Re outputing should not print more data
	outW.Reset()
	errW.Reset()
	err = alphaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, alphaOutputer.hasNext())
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	// Use a second outputer should print all available session datas without global notifications
	out2W.Reset()
	err2W.Reset()
	alphaOutputer2 := zgo.sessionOutputer("alpha")
	require.NotNil(t, alphaOutputer2)
	assert.True(t, alphaOutputer2.hasNext())
	assert.True(t, alphaOutputer2.hasNext())
	err = alphaOutputer2.outputs(outs2)
	assert.NoError(t, err)
	assert.True(t, alphaOutputer2.hasNext())
	assert.Equal(t, "msg-alpha_5_notifier_1\nmsg-alpha_5_notifier_2\nmsg-alpha_5_bar_0_1\n", out2W.String())
	assert.Equal(t, "", err2W.String())

	// bravo session outputs
	outW.Reset()
	errW.Reset()
	bravoOutputer := zgo.sessionOutputer("bravo")
	require.NotNil(t, bravoOutputer)
	assert.True(t, bravoOutputer.hasNext())
	assert.True(t, bravoOutputer.hasNext())
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "msg-bravo_0_notifier_1\nmsg-bravo_0_notifier_2\nmsg-bravo_0_notifier_3\nmsg-bravo_0_pif_0_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Re outputing should not print more data
	outW.Reset()
	errW.Reset()
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = bravoOutputer.update()
	assert.NoError(t, err)
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	// Closing printer should print more data
	outW.Reset()
	errW.Reset()
	err = alphaBarPrinter1.Close("close1")
	assert.NoError(t, err)
	err = alphaOutputer.update()
	assert.NoError(t, err)
	err = alphaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, alphaOutputer.hasNext())
	assert.Equal(t, "msg-alpha_5_bar_0_2\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Closing last printer part should print it first
	outW.Reset()
	errW.Reset()
	err = alphaBarPrinter4.Close("close2")
	assert.NoError(t, err)
	err = alphaBazPrinter1.Close("close3")
	assert.NoError(t, err)
	err = alphaOutputer.update()
	assert.NoError(t, err)
	err = alphaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, alphaOutputer.hasNext())
	assert.Equal(t, "msg-alpha_5_bar_0_4\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Closing all printer parts should print next printer
	// The printer should be considered closed since all parts are closed
	outW.Reset()
	errW.Reset()
	err = alphaBarPrinter5.Close("close4")
	assert.NoError(t, err)
	err = alphaBarPrinter3.Close("close5")
	assert.NoError(t, err)
	err = alphaBarPrinter2.Close("close6")
	assert.NoError(t, err)
	err = alphaOutputer.update()
	assert.NoError(t, err)
	assert.True(t, alphaOutputer.hasNext())
	err = alphaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.False(t, alphaOutputer.hasNext())
	assert.Equal(t, "msg-alpha_5_bar_0_3\nmsg-alpha_5_bar_0_5\nmsg-alpha_5_foo_0_1\n", outW.String())
	assert.Equal(t, "", errW.String())
}

func TestFsLayer_OutputsNotifying(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestFsLayer_OutputsNotifying"
	require.NoError(t, os.RemoveAll(tmpDir))
	filez.MkdirAll(tmpDir, filez.DefaultDirPerms)
	defer os.RemoveAll(tmpDir)
	var err error

	scenario1 := buildTestFsPrinters_scenario1(t, tmpDir)
	globalNotifier1 := scenario1["global_notifier_1"]
	globalNotifier2 := scenario1["global_notifier_2"]
	bravoNotifier2 := scenario1["bravo_0_notifier_2"]

	zg := buildZcreenGroup(tmpDir)
	updated, err := zg.scanFiles()
	assert.True(t, updated)
	require.NoError(t, err)

	zgo := zg.outputer()
	require.NotNil(t, zgo)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)

	// bravo session outputs
	outW.Reset()
	errW.Reset()
	bravoOutputer := zgo.sessionOutputer("bravo")
	require.NotNil(t, bravoOutputer)
	assert.True(t, bravoOutputer.hasNext())
	assert.True(t, bravoOutputer.hasNext())
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "msg-global_notifier_1\nmsg-global_notifier_2\nmsg-global_notifier_3\nmsg-bravo_0_notifier_1\nmsg-bravo_0_notifier_2\nmsg-bravo_0_notifier_3\nmsg-bravo_0_pif_0_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// gamma session outputs
	outW.Reset()
	errW.Reset()
	gammaOutputer := zgo.sessionOutputer("gamma")
	require.NotNil(t, gammaOutputer)
	assert.True(t, gammaOutputer.hasNext())
	assert.True(t, gammaOutputer.hasNext())
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, gammaOutputer.hasNext())
	assert.Equal(t, "msg-gamma_0_baz_10_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Add a first global notif
	globalNotifier1.Out("notif1_more\n")
	err = globalNotifier1.Flush()
	assert.NoError(t, err)

	// First session output display global notif
	outW.Reset()
	errW.Reset()
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "notif1_more\n", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, gammaOutputer.hasNext())
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	// Add a second global notif
	globalNotifier2.Out("notif2_more\n")
	err = globalNotifier2.Flush()
	assert.NoError(t, err)

	// First session output display global notif
	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, gammaOutputer.hasNext())
	assert.Equal(t, "notif2_more\n", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	// Add a session notif
	outW.Reset()
	errW.Reset()
	bravoNotifier2.Out("bravo_notif2_more\n")
	err = bravoNotifier2.Flush()
	assert.NoError(t, err)

	// Only bravo session display the notif
	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, gammaOutputer.hasNext())
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "bravo_notif2_more\n", outW.String())
}

func TestFsLayer_OutputsPrinting(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestFsLayer_OutputsPrinting"
	require.NoError(t, os.RemoveAll(tmpDir))
	filez.MkdirAll(tmpDir, filez.DefaultDirPerms)
	defer os.RemoveAll(tmpDir)
	var err error

	scenario1 := buildTestFsPrinters_scenario1(t, tmpDir)
	bravoPifPrinter1 := scenario1["bravo_0_pif_0_1"]
	bravoPifPrinter3 := scenario1["bravo_0_pif_0_3"]
	bravoPafPrinter1 := scenario1["bravo_0_paf_1_1"]

	zg := buildZcreenGroup(tmpDir)
	updated, err := zg.scanFiles()
	assert.True(t, updated)
	require.NoError(t, err)

	zgo := zg.outputer()
	require.NotNil(t, zgo)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)

	// Todo add prints like add notifications
	// bravo session outputs
	outW.Reset()
	errW.Reset()
	bravoOutputer := zgo.sessionOutputer("bravo")
	require.NotNil(t, bravoOutputer)
	assert.True(t, bravoOutputer.hasNext())
	assert.True(t, bravoOutputer.hasNext())
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "msg-global_notifier_1\nmsg-global_notifier_2\nmsg-global_notifier_3\nmsg-bravo_0_notifier_1\nmsg-bravo_0_notifier_2\nmsg-bravo_0_notifier_3\nmsg-bravo_0_pif_0_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// gamma session outputs
	outW.Reset()
	errW.Reset()
	gammaOutputer := zgo.sessionOutputer("gamma")
	require.NotNil(t, gammaOutputer)
	assert.True(t, gammaOutputer.hasNext())
	assert.True(t, gammaOutputer.hasNext())
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, gammaOutputer.hasNext())
	assert.Equal(t, "msg-gamma_0_baz_10_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Add a first print in first bravo printer => MUST output because it's current outputing printer
	bravoPifPrinter1.Out("bravo_pif_1_more\n")
	err = bravoPifPrinter1.Flush()
	assert.NoError(t, err)

	outW.Reset()
	errW.Reset()
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "bravo_pif_1_more\n", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, gammaOutputer.hasNext())
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	// Add a second print in last bravo printer => MUST NOT output because not current outputing printer
	bravoPifPrinter3.Out("bravo_pif_3_more\n")
	err = bravoPifPrinter3.Flush()
	assert.NoError(t, err)
	bravoPafPrinter1.Out("bravo_paf_1_more\n")
	err = bravoPafPrinter1.Flush()
	assert.NoError(t, err)

	outW.Reset()
	errW.Reset()
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, gammaOutputer.hasNext())
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	// Close first & last printers => MUST output last printer data then middle printer not closed data
	err = bravoPifPrinter3.Close("msg3")
	assert.NoError(t, err)
	err = bravoPifPrinter1.Close("msg1")
	assert.NoError(t, err)
	err = bravoOutputer.update()
	assert.NoError(t, err)

	outW.Reset()
	errW.Reset()
	err = bravoOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, bravoOutputer.hasNext())
	assert.Equal(t, "msg-bravo_0_pif_0_3\nbravo_pif_3_more\nmsg-bravo_0_pif_0_2\n", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.True(t, gammaOutputer.hasNext())
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())
}

func TestFsLayer_AddingPrinters(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestFsLayer_AddingPrinters"
	require.NoError(t, os.RemoveAll(tmpDir))
	filez.MkdirAll(tmpDir, filez.DefaultDirPerms)
	defer os.RemoveAll(tmpDir)
	var err error

	scenario1 := buildTestFsPrinters_scenario1(t, tmpDir)
	gammaBazPrinter1 := scenario1["gamma_0_baz_10_1"]
	gammaBarPrinter1 := scenario1["gamma_0_bar_20_1"]

	zg := buildZcreenGroup(tmpDir)
	updated, err := zg.scanFiles()
	assert.True(t, updated)
	require.NoError(t, err)

	zgo := zg.outputer()
	require.NotNil(t, zgo)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)

	gammaOutputer := zgo.sessionOutputer("gamma")
	require.NotNil(t, gammaOutputer)
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "msg-global_notifier_1\nmsg-global_notifier_2\nmsg-global_notifier_3\nmsg-gamma_0_baz_10_1\n", outW.String())
	assert.Empty(t, errW.String())

	// add new printers with higher same & lower priority
	_, _, gammaNewHighPrioPrinter := buildTestFsPrinter(t, &scenario1, tmpDir, "gamma", 0, "high", 5, 1)
	_, _, gammaNewSamePrioPrinter := buildTestFsPrinter(t, &scenario1, tmpDir, "gamma", 0, "same", 10, 1)
	_, _, gammaNewLowPrioPrinter := buildTestFsPrinter(t, &scenario1, tmpDir, "gamma", 0, "low", 15, 1)
	_ = gammaNewHighPrioPrinter
	_ = gammaNewSamePrioPrinter
	_ = gammaNewLowPrioPrinter
	//gammaNewHighPrioPrinter.Out("high_prio_1\n")
	//err = gammaNewHighPrioPrinter.Flush()
	//assert.NoError(t, err)
	//gammaNewSamePrioPrinter.Out("same_prio_1\n")
	//err = gammaNewSamePrioPrinter.Flush()
	//assert.NoError(t, err)
	//gammaNewLowPrioPrinter.Out("low_prio_1\n")
	//err = gammaNewLowPrioPrinter.Flush()
	//assert.NoError(t, err)
	err = gammaOutputer.update()
	assert.NoError(t, err)

	// First nothing MUST be displayed while current printer not closed
	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	// Second only printer with same priority MUST be outputed
	err = gammaBazPrinter1.Close("msg1")
	assert.NoError(t, err)
	err = gammaOutputer.update()
	assert.NoError(t, err)
	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "msg-gamma_0_same_10_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Third only printer with higher priority MUST be outputed
	err = gammaNewSamePrioPrinter.Close("msg2")
	assert.NoError(t, err)
	err = gammaOutputer.update()
	assert.NoError(t, err)
	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "msg-gamma_0_high_5_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Eventually lower priority printer MUST be displayed
	err = gammaNewHighPrioPrinter.Close("msg3")
	assert.NoError(t, err)
	err = gammaOutputer.update()
	assert.NoError(t, err)
	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "msg-gamma_0_low_15_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Initial lowest priority printer MUST be displayed
	err = gammaNewLowPrioPrinter.Close("msg4")
	assert.NoError(t, err)
	err = gammaOutputer.update()
	assert.NoError(t, err)
	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "msg-gamma_0_bar_20_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// After everything is closed nothing left to output. all printer groups SHOULD be closed, session group SHOULD not be ended.
	assert.True(t, gammaOutputer.hasNext()) // has next because one printer still opened
	err = gammaBarPrinter1.Close("msg5")
	assert.NoError(t, err)
	err = gammaOutputer.update()
	assert.NoError(t, err)
	assert.True(t, gammaOutputer.hasNext()) // has next because last printer need to be outputed
	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())
	assert.False(t, gammaOutputer.hasNext()) // session fully outputted

	gammaSg := *zg.sessionsByPrioName[forgePrioNameKey(0, "gamma")]
	require.NotNil(t, gammaSg)
	gamma_bar_pg := gammaSg.printersByPrioName[forgePrioNameKey(20, "bar")]
	require.NotNil(t, gamma_bar_pg)
	gamma_baz_pg := gammaSg.printersByPrioName[forgePrioNameKey(10, "baz")]
	require.NotNil(t, gamma_baz_pg)
	gamma_high_pg := gammaSg.printersByPrioName[forgePrioNameKey(5, "high")]
	require.NotNil(t, gamma_high_pg)
	gamma_same_pg := gammaSg.printersByPrioName[forgePrioNameKey(10, "same")]
	require.NotNil(t, gamma_same_pg)
	gamma_low_pg := gammaSg.printersByPrioName[forgePrioNameKey(15, "low")]
	require.NotNil(t, gamma_low_pg)

	assert.False(t, gammaSg.ended)
	assert.True(t, gamma_bar_pg.closed)
	assert.True(t, gamma_baz_pg.closed)
	assert.True(t, gamma_high_pg.closed)
	assert.True(t, gamma_same_pg.closed)
	assert.True(t, gamma_low_pg.closed)

	// Adding a new printer in an opened session with all printer closed
	buildTestFsPrinter(t, &scenario1, tmpDir, "gamma", 0, "extra", 0, 1)
	err = gammaOutputer.update()
	assert.NoError(t, err)
	outW.Reset()
	errW.Reset()
	err = gammaOutputer.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "msg-gamma_0_extra_0_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Adding a printer in an ended session must not be outputed
	// TODO: From FS side do we need to block printers on ended session ?
	// TODO: ?
}

func TestFsLayer_SharingAndOrderingPrinters(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestFsLayer_SharingAndOrderingPrinters"
	require.NoError(t, os.RemoveAll(tmpDir))
	filez.MkdirAll(tmpDir, filez.DefaultDirPerms)
	//defer os.RemoveAll(tmpDir)
	var err error

	uniqSession1 := "paf47001-uniq1"
	uniqSession2 := "paf47001-uniq2"
	sharedSession := "paf47001-shared"

	// ----- First zcreenGroup
	zg1 := buildZcreenGroup(tmpDir)
	require.NotNil(t, zg1)
	updated, err := zg1.scanFiles()
	assert.NoError(t, err)
	assert.False(t, updated)
	zg1o := zg1.outputer()
	require.NotNil(t, zg1o)
	u1so := zg1o.sessionOutputer(uniqSession1)
	require.NotNil(t, u1so)
	s0so1 := zg1o.sessionOutputer(sharedSession)
	require.NotNil(t, s0so1)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)

	m := make(map[string]*fsPrinter)

	// ----- zcreen 1 write tests
	// Write on uniq printer
	_, _, u1_prtr_10 := buildTestFsPrinter(t, &m, tmpDir, uniqSession1, 0, "u1", 10, 1)
	require.NotNil(t, u1_prtr_10)
	// assert.NotPanics(t, func() {
	// 	u1_prtr_10.Out("u1_prtr_10\n")
	// })

	// Write on shared printer
	_, _, s0_prtr_10 := buildTestFsPrinter(t, &m, tmpDir, sharedSession, 0, "s0", 10, 1)
	require.NotNil(t, s0_prtr_10)
	// assert.NotPanics(t, func() {
	// 	s0_prtr_10.Out("s0_prtr_10\n")
	// })

	_, _, s0_prtr_50 := buildTestFsPrinter(t, &m, tmpDir, sharedSession, 0, "s0", 50, 1)
	require.NotNil(t, s0_prtr_50)
	// assert.NotPanics(t, func() {
	// 	s0_prtr_50.Out("s0_prtr_50\n")
	// })

	_, _, s0_prtr_30a := buildTestFsPrinter(t, &m, tmpDir, sharedSession, 0, "s0", 30, 1)
	require.NotNil(t, s0_prtr_30a)
	// assert.NotPanics(t, func() {
	// 	s0_prtr_30a.Out("s0_prtr_30a\n")
	// })

	err = s0_prtr_10.Close("foo")
	assert.NoError(t, err)
	err = u1_prtr_10.Close("foo")
	assert.NoError(t, err)
	err = s0_prtr_50.Close("foo")
	assert.NoError(t, err)
	err = s0_prtr_30a.Close("foo")
	assert.NoError(t, err)

	// ----- Second screen
	zg2 := buildZcreenGroup(tmpDir)
	require.NotNil(t, zg2)
	// updated, err = zg2.scanFiles()
	// assert.NoError(t, err)
	// assert.True(t, updated)
	zg2o := zg2.outputer()
	require.NotNil(t, zg2o)
	u2so := zg2o.sessionOutputer(uniqSession2)
	require.NotNil(t, u2so)
	s0so2 := zg2o.sessionOutputer(sharedSession)
	require.NotNil(t, s0so2)

	// ----- zcreen 2 write tests
	// Write on uniq printer
	_, _, u2_prtr_10 := buildTestFsPrinter(t, &m, tmpDir, uniqSession2, 0, "u2", 10, 1)
	require.NotNil(t, u2_prtr_10)
	// assert.NotPanics(t, func() {
	// 	u2_prtr_10.Out("u2_prtr_10\n")
	// })

	// Write on shared printer
	_, _, s0_prtr_30b := buildTestFsPrinter(t, &m, tmpDir, sharedSession, 0, "s0", 30, 2)
	require.NotNil(t, s0_prtr_30b)
	// assert.NotPanics(t, func() {
	// 	s0_prtr_30b.Out("s0_prtr_30b\n")
	// })

	_, _, s0_prtr_20 := buildTestFsPrinter(t, &m, tmpDir, sharedSession, 0, "s0", 20, 1)
	require.NotNil(t, s0_prtr_20)
	// assert.NotPanics(t, func() {
	// 	s0_prtr_20.Out("s0_prtr_20\n")
	// })

	_, _, s0_prtr_40 := buildTestFsPrinter(t, &m, tmpDir, sharedSession, 0, "s0", 40, 1)
	require.NotNil(t, s0_prtr_40)
	// assert.NotPanics(t, func() {
	// 	s0_prtr_40.Out("s0_prtr_40\n")
	// })

	err = s0_prtr_30b.Close("foo")
	assert.NoError(t, err)
	err = s0_prtr_40.Close("foo")
	assert.NoError(t, err)
	err = u2_prtr_10.Close("foo")
	assert.NoError(t, err)
	err = s0_prtr_20.Close("foo")
	assert.NoError(t, err)

	// ----- Outputing session by session

	// Check zcreen group content: FS not scanned => nothing to output.
	outW.Reset()
	errW.Reset()
	err = u1so.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = s0so1.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = u2so.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = s0so2.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	// Scan one zcreen group: Only one should output
	err = u1so.update()
	assert.NoError(t, err)
	err = s0so1.update()
	assert.NoError(t, err)

	outW.Reset()
	errW.Reset()
	err = u1so.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "msg-paf47001-uniq1_0_u1_10_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = s0so1.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "msg-paf47001-shared_0_s0_10_1\nmsg-paf47001-shared_0_s0_20_1\nmsg-paf47001-shared_0_s0_30a_1\nmsg-paf47001-shared_0_s0_30b_2\nmsg-paf47001-shared_0_s0_40_1\nmsg-paf47001-shared_0_s0_50_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = u2so.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = s0so2.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	// Scan second zcreen group
	err = u2so.update()
	assert.NoError(t, err)
	err = s0so2.update()
	assert.NoError(t, err)

	outW.Reset()
	errW.Reset()
	err = u1so.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = s0so1.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = u2so.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "msg-paf47001-uniq2_0_u2_10_1\n", outW.String())
	assert.Equal(t, "", errW.String())

	outW.Reset()
	errW.Reset()
	err = s0so2.outputs(outs)
	assert.NoError(t, err)
	assert.Equal(t, "msg-paf47001-shared_0_s0_10_1\nmsg-paf47001-shared_0_s0_20_1\nmsg-paf47001-shared_0_s0_30a_1\nmsg-paf47001-shared_0_s0_30b_2\nmsg-paf47001-shared_0_s0_40_1\nmsg-paf47001-shared_0_s0_50_1\n", outW.String())
	assert.Equal(t, "", errW.String())

}
