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




