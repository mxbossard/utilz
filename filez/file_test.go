package filez

import (
	"bytes"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

var ()

func TestMkTemp(t *testing.T) {
	expected := "filez.foo10."
	path, err := MkTemp(expected)
	assert.NoError(t, err)
	assert.Contains(t, path, "/"+expected)
	defer os.RemoveAll(path)

	_, err = os.Stat(path)
	assert.Error(t, err)
	assert.ErrorIs(t, err, fs.ErrNotExist)

	_, err = MkTemp("/")
	assert.Error(t, err)
}

func TestMkTempOrPanic(t *testing.T) {
	assert.Panics(t, func() {
		MkTempOrPanic("/")
	})
}

func TestMkTemp2(t *testing.T) {
	expected := "filez.foo20."
	expectedDir := "filez.dir20."
	dir, err := os.MkdirTemp("", expectedDir)
	assert.NoError(t, err)
	assert.Contains(t, dir, "/"+expectedDir)

	path, err := MkTemp2(dir, expected)
	assert.NoError(t, err)
	assert.Contains(t, path, "/"+expectedDir)
	assert.Contains(t, path, "/"+expected)

	_, err = MkTemp2("-do-not-exists", expectedDir)
	assert.Error(t, err)
}

func TestMkTemp2OrPanic(t *testing.T) {
	expectedDir := "filez.dir20."
	assert.Panics(t, func() {
		MkTemp2OrPanic("-do-not-exists", expectedDir)
	})
}

func TestOpenTemp(t *testing.T) {
	expected := "filez.foo30."
	f, err := OpenTemp(expected)
	assert.NoError(t, err)

	info, err := f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, DefaultFilePerms, info.Mode())
	assert.False(t, info.IsDir())
	assert.Contains(t, info.Name(), expected)
	assert.Greater(t, len(info.Name()), len(expected))

	_, err = OpenTemp("/")
	assert.Error(t, err)
}

func TestOpenTempOrPanic(t *testing.T) {
	assert.Panics(t, func() {
		OpenTempOrPanic("/")
	})
}

func TestMkdirTemp(t *testing.T) {
	expected := "filez.bar."
	path, err := MkdirTemp(expected)
	assert.NoError(t, err)
	assert.Contains(t, path, expected)
	defer os.RemoveAll(path)

	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, DefaultDirPerms, info.Mode().Perm())
	assert.True(t, info.IsDir())
	assert.Contains(t, info.Name(), expected)
	assert.Greater(t, len(info.Name()), len(expected))

	_, err = MkdirTemp("/")
	assert.Error(t, err)
}

func TestMkdirTempOrPanic(t *testing.T) {
	assert.Panics(t, func() {
		MkdirTempOrPanic("/")
	})
}

func TestMkdirTemp2(t *testing.T) {
	// TODO: test existing dir 
	// TODO: test not existing dir

	_, err := MkdirTemp2("", "/")
	assert.Error(t, err)
}

func TestMkdirTemp2OrPanic(t *testing.T) {
	assert.Panics(t, func() {
		MkdirTemp2OrPanic("", "/")
	})
}

func TestOpen(t *testing.T) {
	// TODO

	_, err := Open("/-do-not-exists-_")
	assert.Error(t, err)
}

func TestOpenOrPanic(t *testing.T) {
	assert.Panics(t, func() {
		OpenOrPanic("/-do-not-exists-_")
	})
}

func TestOpen3(t *testing.T) {
	// TODO

	_, err := Open3("/-do-not-exists-_", 42, 0400)
	assert.Error(t, err)
}

func TestOpen3OrPanic(t *testing.T) {
	assert.Panics(t, func() {
		Open3OrPanic("/-do-not-exists-_", 42, 0400)
	})
}

func TestMkdirAll(t *testing.T) {
	// TODO

	err := MkdirAll("", 0000)
	assert.Error(t, err)
}

func TestMkdirAllOrPanic(t *testing.T) {
	assert.Panics(t, func() {
		MkdirAllOrPanic("", 0400)
	})
}

func TestMkdir(t *testing.T) {
	// TODO

	err := Mkdir("", 0400)
	assert.Error(t, err)
}

func TestMkdirOrPanic(t *testing.T) {
	assert.Panics(t, func() {
		MkdirOrPanic("", 0400)
	})
}


func TestChdirAndWorkingDir(t *testing.T) {
	// TODO
}

func TestChdirOrPanic(t *testing.T) {
	assert.Panics(t, func() {
		ChdirOrPanic("")
	})
}

func TestWorkingDirOrPanic(t *testing.T) {
	assert.Panics(t, func() {
		// TODO: how ?
		panic("TODO: how test it ?")
	})
}

func TestMkSubDirAll(t *testing.T) {
	// TODO
}

func TestMkSubDirAll3(t *testing.T) {
	// TODO
}

func TestMkSubDir(t *testing.T) {
	// TODO
}

func TestMkSubDir3(t *testing.T) {
	// TODO
}

func TestSoftInitFile(t *testing.T) {
	// TODO
}

func TestSoftInitFile3(t *testing.T) {
	// TODO
}

func TestWriteStringThenReadString(t *testing.T) {
	path, err := MkTemp("filez.baz.*")
	assert.NoError(t, err)
	defer os.RemoveAll(path)

	expected := "mymessage\nfoobar"
	var expectedPerms fs.FileMode = 0604
	err = WriteString(path, expected, expectedPerms)
	assert.NoError(t, err)
	s, err := ReadString(path)
	assert.NoError(t, err)
	assert.Equal(t, expected, s)

	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, expectedPerms, info.Mode().Perm())
}

func TestReadOrPanic(t *testing.T) {
	assert.Panics(t, func() {
	// TODO
		panic("TODO")
	})
}

func TestReadStringOrPanic(t *testing.T) {
	assert.Panics(t, func() {
	// TODO
		panic("TODO")
	})
}

func TestWriteOrPanic(t *testing.T) {
	assert.Panics(t, func() {
	// TODO
		panic("TODO")
	})
}

func TestWriteStringOrPanic(t *testing.T) {
	assert.Panics(t, func() {
	// TODO
		panic("TODO")
	})
}

func TestReadFile(t *testing.T) {
	// TODO
}

func TestReadFileOrPanic(t *testing.T) {
	assert.Panics(t, func() {
	// TODO
		panic("TODO")
	})
}

func TestReadFileString(t *testing.T) {
	// TODO
}

func TestReadFileStringOrPanic(t *testing.T) {
	assert.Panics(t, func() {
	// TODO
		panic("TODO")
	})
}

func TestPrint(t *testing.T) {
	// TODO
}

func TestPrintOrPanic(t *testing.T) {
	assert.Panics(t, func() {
	// TODO
		panic("TODO")
	})
}

func TestPrintTree(t *testing.T) {
	// TODO
}

func TestPrintTreeOrPanic(t *testing.T) {
	assert.Panics(t, func() {
	// TODO
		panic("TODO")
	})
}

func TestIsDirectory(t *testing.T) {
	// TODO
}

func TestIsDirectoryOrPanic(t *testing.T) {
	assert.Panics(t, func() {
	// TODO
		panic("TODO")
	})
}

func TestCopy_SmallBuffer(t *testing.T) {
	src, err := MkTemp("filez.src.*")
	assert.NoError(t, err)
	defer os.RemoveAll(src)
	expected := "mymessage\nfoobar"
	err = WriteString(src, expected, 0600)
	assert.NoError(t, err)
	assert.Equal(t, expected, ReadStringOrPanic(src))
	
	dst := &bytes.Buffer{}

	b := make([]byte, 3)
	fSrc, err := Open(src)
	assert.NoError(t, err)
	n, err := Copy(fSrc, dst, b)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(expected)), n)

	assert.Equal(t, expected, dst.String())
}

func TestCopy_LargeBuffer(t *testing.T) {
	src, err := MkTemp("filez.src.*")
	assert.NoError(t, err)
	defer os.RemoveAll(src)
	expected := "mymessage\nfoobar"
	err = WriteString(src, expected, 0600)
	assert.NoError(t, err)
	assert.Equal(t, expected, ReadStringOrPanic(src))
	
	dst := &bytes.Buffer{}

	b := make([]byte, len(expected) * 10)
	fSrc, err := Open(src)
	assert.NoError(t, err)
	n, err := Copy(fSrc, dst, b)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(expected)), n)

	assert.Equal(t, expected, dst.String())
}

func TestCopyChunk_UntilEnd(t *testing.T) {
	src, err := MkTemp("filez.src.*")
	assert.NoError(t, err)
	defer os.RemoveAll(src)
	expected := "mymessage\nfoobar"
	err = WriteString(src, expected, 0600)
	assert.NoError(t, err)
	assert.Equal(t, expected, ReadStringOrPanic(src))
	
	dst := &bytes.Buffer{}

	b := make([]byte, 3)
	fSrc, err := Open(src)
	assert.NoError(t, err)
	n, err := CopyChunk(fSrc, dst, b, 7, -1)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(expected)-7), n)
	assert.Equal(t, expected[7:], dst.String())
}

func TestCopyChunk_SmallBuffer(t *testing.T) {
	src, err := MkTemp("filez.src.*")
	assert.NoError(t, err)
	defer os.RemoveAll(src)
	expected := "mymessage\nfoobar"
	err = WriteString(src, expected, 0600)
	assert.NoError(t, err)
	assert.Equal(t, expected, ReadStringOrPanic(src))
	
	dst := &bytes.Buffer{}

	b := make([]byte, 3)
	fSrc, err := Open(src)
	assert.NoError(t, err)
	n, err := CopyChunk(fSrc, dst, b, 7, 12)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), n)
	assert.Equal(t, expected[7:12], dst.String())
}

func TestCopyChunk_LargeBuffer(t *testing.T) {
	src, err := MkTemp("filez.src.*")
	assert.NoError(t, err)
	defer os.RemoveAll(src)
	expected := "mymessage\nfoobar"
	err = WriteString(src, expected, 0600)
	assert.NoError(t, err)
	assert.Equal(t, expected, ReadStringOrPanic(src))
	
	dst := &bytes.Buffer{}

	b := make([]byte, len(expected) * 10)
	fSrc, err := Open(src)
	assert.NoError(t, err)
	n, err := CopyChunk(fSrc, dst, b, 7, 12)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), n)
	assert.Equal(t, expected[7:12], dst.String())
}




