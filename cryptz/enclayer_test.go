package cryptz

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/truzt"
	"github.com/stretchr/testify/assert"
)

const (
	dummyNamePrefix = "pre"
	dummyNameSuffix = "suf"
	dummyEncPrefix  = "foo"
	dummyEncSuffix  = "bar"
)

func dummyEncryptionFactory() encryptionFactory {
	return encryptionFactory{
		namer: DummyNamer{
			prefix: dummyNamePrefix,
			suffix: dummyNameSuffix,
		},
		encrypter: DummyEncrypter{
			prefix: dummyEncPrefix,
			suffix: dummyEncSuffix,
		},
	}
}

func TestDumyOpenFileScenario(t *testing.T) {
	// Test encryption factory without encryption
	pubDir := filez.MkdirTempOrPanic("TestDumyOpenFileScenario")
	defer os.RemoveAll(pubDir)

	pubDir = filepath.Join(pubDir, "pubDir")
	privPath := filepath.Join("dir1", "dir2", "file")
	expectedMsg := "msg"

	ef := dummyEncryptionFactory()

	// --- Open a file an write into it
	f, err := ef.OpenFile(nil, pubDir, privPath, os.O_CREATE+os.O_RDWR, 0600)
	assert.NoError(t, err)
	assert.FileExists(t, f.Name())
	assert.Equal(t, filepath.Join(pubDir, dummyNamePrefix+"dir1"+dummyNameSuffix, dummyNamePrefix+"dir2"+dummyNameSuffix, dummyNamePrefix+"file"+dummyNameSuffix), f.Name())

	n, err := f.WriteString(expectedMsg)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedMsg), n)

	err = f.Sync()
	assert.NoError(t, err)

	assert.Equal(t, dummyEncPrefix+expectedMsg+dummyEncSuffix, filez.ReadAtStringOrPanic(f.Name(), nonceSize))

	err = f.Close()
	assert.NoError(t, err)

	assert.Equal(t, dummyEncPrefix+expectedMsg+dummyEncSuffix, filez.ReadAtStringOrPanic(f.Name(), nonceSize))

	// --- ReOpen the file ReadOnly and check content
	g, err := ef.Open(nil, pubDir, privPath)
	assert.NoError(t, err)
	assert.FileExists(t, g.Name())
	assert.Equal(t, f.Name(), g.Name())

	sw := bytes.NewBuffer(make([]byte, 0, 1000))
	p, err := g.WriteTo(sw)
	assert.NoError(t, err)
	assert.Equal(t, int64(len([]byte(expectedMsg))), p)
	assert.Equal(t, expectedMsg, sw.String())

	_, err = g.WriteString(expectedMsg)
	assert.Error(t, err)

	err = g.Close()
	assert.NoError(t, err)
}

func TestDefaultImplOpenFileScenario(t *testing.T) {
	// Test encryption factory file opening write and readonly
	pubDir := filez.MkdirTempOrPanic("TestDefaultImplOpenFileScenario")
	defer os.RemoveAll(pubDir)

	pubDir = filepath.Join(pubDir, "pubDir")
	privPath := filepath.Join("dir1", "dir2", "file")
	expectedMsg := "msg"

	kh := NewStringKeyHolder("pif", "paf")
	ef := NewDefaultEncryptionFactory(kh)

	// --- Open a file an write into it
	f, err := ef.OpenFile(nil, pubDir, privPath, os.O_CREATE+os.O_RDWR, 0600)
	assert.NoError(t, err)
	assert.FileExists(t, f.Name())
	// assert.Equal(t, filepath.Join(pubDir, dummyNamePrefix+"dir1"+dummyNameSuffix, dummyNamePrefix+"dir2"+dummyNameSuffix, dummyNamePrefix+"file"+dummyNameSuffix), f.Name())

	n, err := f.WriteString(expectedMsg)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedMsg), n)

	err = f.Sync()
	assert.NoError(t, err)

	// assert.Equal(t, dummyEncPrefix+expectedMsg+dummyEncSuffix, filez.ReadStringOrPanic(f.Name()))

	err = f.Close()
	assert.NoError(t, err)

	// assert.Equal(t, dummyEncPrefix+expectedMsg+dummyEncSuffix, filez.ReadStringOrPanic(f.Name()))

	// --- ReOpen the file ReadOnly and check content
	g, err := ef.Open(nil, pubDir, privPath)
	assert.NoError(t, err)
	assert.FileExists(t, g.Name())
	assert.Equal(t, f.Name(), g.Name())

	sw := bytes.NewBuffer(make([]byte, 0, 1000))
	p, err := g.WriteTo(sw)
	assert.NoError(t, err)
	assert.Equal(t, int64(len([]byte(expectedMsg))), p)
	assert.Equal(t, expectedMsg, sw.String())

	_, err = g.WriteString(expectedMsg)
	assert.Error(t, err)

	err = g.Close()
	assert.NoError(t, err)
}

func TestDefaultImplSuccessivesEncryptionsScenario(t *testing.T) {
	// Test encryption factory file write, rewrite without updated & updates.
	// Resync a not modified file should not modify the encrypted file.
	// Successives encryption must always lead to different ciphertext.
	pubDir := filez.MkdirTempOrPanic("TestDefaultImplSuccessivesEncryptionsScenario")
	defer os.RemoveAll(pubDir)

	pubDir = filepath.Join(pubDir, "pubDir")
	privPath := filepath.Join("dir1", "dir2", "file")
	expectedMsg1 := "msg"
	expectedMsg2 := "msg2"

	kh := NewStringKeyHolder("pif", "paf")
	ef := NewDefaultEncryptionFactory(kh)

	// --- Open a file an write into it
	f, err := ef.OpenFile(nil, pubDir, privPath, os.O_CREATE+os.O_RDWR, 0600)
	assert.NoError(t, err)
	assert.FileExists(t, f.Name())
	// assert.Equal(t, filepath.Join(pubDir, dummyNamePrefix+"dir1"+dummyNameSuffix, dummyNamePrefix+"dir2"+dummyNameSuffix, dummyNamePrefix+"file"+dummyNameSuffix), f.Name())

	n, err := f.WriteString(expectedMsg1)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedMsg1), n)

	err = f.Sync()
	assert.NoError(t, err)

	ciphertext1 := filez.ReadStringOrPanic(f.Name())
	sign1, err := truzt.SignFile(f.Name())
	assert.NoError(t, err)
	metadata1, err := truzt.SerializeFileInfo(f)
	assert.NoError(t, err)
	assert.NotEqual(t, expectedMsg1, ciphertext1)

	// Not updating file should not lead to file changes
	err = f.Sync()
	assert.NoError(t, err)

	ciphertext2 := filez.ReadStringOrPanic(f.Name())
	sign2, err := truzt.SignFile(f.Name())
	assert.NoError(t, err)
	metadata2, err := truzt.SerializeFileInfo(f)
	assert.NoError(t, err)
	assert.Equal(t, ciphertext1, ciphertext2)
	assert.Equal(t, sign1, sign2)
	assert.Equal(t, string(metadata1), string(metadata2))

	// Updating file should lead to file changes
	n, err = f.WriteString(expectedMsg2)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedMsg2), n)
	err = f.Sync()
	assert.NoError(t, err)

	ciphertext3 := filez.ReadStringOrPanic(f.Name())
	assert.NotEqual(t, expectedMsg2, ciphertext3)
	assert.NotEqual(t, ciphertext1, ciphertext3)

	// Going back to first message should lead to file changes
	n, err = f.WriteString(expectedMsg1)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedMsg1), n)
	err = f.Sync()
	assert.NoError(t, err)

	ciphertext4 := filez.ReadStringOrPanic(f.Name())
	assert.NotEqual(t, expectedMsg2, ciphertext4)
	assert.NotEqual(t, ciphertext1, ciphertext4)
	assert.NotEqual(t, ciphertext3, ciphertext4)

	err = f.Close()
	assert.NoError(t, err)
}
