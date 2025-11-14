package filez

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/mxbossard/utilz/errorz"
)

const (
	DefaultDirPerms  = fs.FileMode(0700)
	DefaultFilePerms = fs.FileMode(0600)
)

func manageError(err error) bool {
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return false
	}
	return true
}

/** Return a temp file path. Do not touch the file. */
func MkTemp(pattern string) (string, error) {
	dir := os.TempDir()
	return MkTemp2(dir, pattern)
}

/** Return a temp file path. Do not touch the file. */
func MkTempOrPanic(pattern string) string {
	f, err := MkTemp(pattern)
	if err != nil {
		panic(err)
	}
	return f
}

/** Return a temp file path. Do not touch the file. */
func MkTemp2(dir, pattern string) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	path := f.Name()
	// FIXME: should not create/touch/open the file in first place !
	defer func() { err = os.Remove(path) }()
	return path, err
}

/** Return a temp file path. Do not touch the file. */
func MkTemp2OrPanic(dir, pattern string) string {
	f, err := MkTemp2(dir, pattern)
	if err != nil {
		panic(err)
	}
	return f
}

func OpenTemp(pattern string) (*os.File, error) {
	f, err := os.CreateTemp("", pattern)
	return f, err
}

func OpenTempOrPanic(pattern string) *os.File {
	f, err := OpenTemp(pattern)
	if err != nil {
		panic(err)
	}
	return f
}

func MkdirTemp(pattern string) (string, error) {
	dir := os.TempDir()
	p, err := MkdirTemp2(dir, pattern)
	return p, err
}

func MkdirTempOrPanic(pattern string) string {
	p, err := MkdirTemp(pattern)
	if err != nil {
		panic(err)
	}
	return p
}

func MkdirTemp2(dir, pattern string) (string, error) {
	p, err := os.MkdirTemp(dir, pattern)
	return p, err
}

func MkdirTemp2OrPanic(dir, pattern string) string {
	p, err := MkdirTemp2(dir, pattern)
	if err != nil {
		panic(err)
	}
	return p
}

func Open(name string) (*os.File, error) {
	return os.Open(name)
}

func OpenOrPanic(name string) *os.File {
	f, err := Open(name)
	if err != nil {
		panic(err)
	}
	return f
}

func Touch(name string) (*os.File, error) {
	return Open3(name, os.O_CREATE, DefaultFilePerms)
}

func TouchOrPanic(name string) *os.File {
	f, err := Open3(name, os.O_CREATE, DefaultFilePerms)
	if err != nil {
		panic(err)
	}
	return f
}

func Open3(name string, flag int, perm fs.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func Open3OrPanic(name string, flag int, perm fs.FileMode) *os.File {
	f, err := Open3(name, flag, perm)
	if err != nil {
		panic(err)
	}
	return f
}

// Create a directory. It may already exists.
// FIXME: remove ?
func createDirectory(path string) (err error) {
	err = os.MkdirAll(path, DefaultDirPerms)
	return
}

func MkdirAll(path string, perm fs.FileMode) error {
	err := os.MkdirAll(path, perm)
	return err
}

func MkdirAllOrPanic(path string, perm fs.FileMode) {
	err := MkdirAll(path, perm)
	if err != nil {
		panic(err)
	}
}

// Create a new directory.
// FIXME: remove ?
func createNewDirectory(path string) (err error) {
	err = os.Mkdir(path, DefaultDirPerms)
	return
}

func Mkdir(path string, perm fs.FileMode) error {
	err := os.Mkdir(path, perm)
	return err
}

func MkdirOrPanic(path string, perm fs.FileMode) {
	err := Mkdir(path, perm)
	if err != nil {
		panic(err)
	}
}

func Chdir(path string) (err error) {
	err = os.Chdir(path)
	return
}

func ChdirOrPanic(path string) {
	err := os.Chdir(path)
	if err != nil {
		panic(err)
	}
}

// Get working directory path.
// Fail if cannot get working directory.
// FIXME: remove ?
func workDirPath() (path string, err error) {
	path, err = os.Getwd()
	return
}

func WorkingDir() (path string, err error) {
	path, err = os.Getwd()
	return
}

func WorkingDirOrPanic() string {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return path
}

func MkSubDirAll(parentDirPath, name string) (string, error) {
	return MkSubDirAll3(parentDirPath, name, DefaultDirPerms)
}

func MkSubDirAll3(parentDirPath, name string, perm fs.FileMode) (path string, err error) {
	path = filepath.Join(parentDirPath, name)
	err = MkdirAll(path, perm)
	return
}

func MkSubDir(parentDirPath, name string) (string, error) {
	return MkSubDir3(parentDirPath, name, DefaultDirPerms)
}

func MkSubDir3(parentDirPath, name string, perm fs.FileMode) (path string, err error) {
	path = filepath.Join(parentDirPath, name)
	err = Mkdir(path, perm)
	return

}

func SoftInitFile(filepath, content string) (string, error) {
	return SoftInitFile3(filepath, content, DefaultFilePerms)
}

func SoftInitFile3(filepath, content string, perm fs.FileMode) (path string, err error) {
	_, err = os.Stat(filepath)
	if os.IsNotExist(err) {
		// Do not overwrite file if it already exists
		err = os.WriteFile(filepath, []byte(content), perm)
	}
	return
}

func Read(filepath string) (content []byte, err error) {
	content, err = os.ReadFile(filepath)
	return
}

func ReadOrPanic(filepath string) (content []byte) {
	content, err := Read(filepath)
	if err != nil {
		panic(err)
	}
	return
}

func ReadString(filepath string) (content string, err error) {
	var bytes []byte
	bytes, err = Read(filepath)
	if err != nil {
		return
	}
	content = string(bytes)
	return
}

func ReadStringOrPanic(filepath string) string {
	content, err := ReadString(filepath)
	if err != nil {
		panic(err)
	}
	return content
}

func WriteString(filepath, content string, perm os.FileMode) (err error) {
	bytes := []byte(content)
	err = os.WriteFile(filepath, bytes, perm)
	return
}

func WriteStringOrPanic(filepath, content string, perm os.FileMode) {
	err := WriteString(filepath, content, perm)
	if err != nil {
		panic(err)
	}
}

func ReadFile(f *os.File, maxLen int) ([]byte, error) {
	b := make([]byte, maxLen)
	n, err := f.ReadAt(b, 0)
	return b[:n], err
}

func ReadFileOrPanic(f *os.File, maxLen int) []byte {
	b := make([]byte, maxLen)
	n, err := f.ReadAt(b, 0)
	if err != nil {
		panic(err)
	}
	return b[:n]
}

func ReadFileString(f *os.File, maxLen int) (string, error) {
	b, err := ReadFile(f, maxLen)
	if err != nil {
		return "", err
	}
	return string(b), err
}

func ReadFileStringOrPanic(f *os.File, maxLen int) string {
	return string(ReadFileOrPanic(f, maxLen))
}

func Print(filepath string) (err error) {
	content, err := Read(filepath)
	if err != nil {
		return
	}
	fmt.Printf("\n%s\n", content)
	return
}

func PrintOrPanic(filepath string) {
	err := Print(filepath)
	if err != nil {
		panic(err)
	}
}

func PrintTree(parentPath string) error {
	err := filepath.Walk(parentPath, func(name string, info os.FileInfo, err error) error {
		fmt.Println(name)
		return nil
	})
	return err
}

func PrintTreeOrPanic(parentPath string) {
	err := PrintTree(parentPath)
	if err != nil {
		panic(err)
	}
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
func ExistsOrPanic(path string) bool {
	ok, err := Exists(path)
	if err != nil {
		panic(err)
	}
	return ok
}

func WaitUntilExists(path string, timeout time.Duration) error {
	start := time.Now()
	for {
		if time.Since(start) > timeout {
			return errorz.Timeoutf(timeout, "waiting file: %s to exists", path)
		}
		ok, err := Exists(path)
		if err != nil {
			return err
		}
		if ok {
			break
		}
		time.Sleep(100 * time.Microsecond)
	}
	return nil
}

func WaitUntilExistsOrPanic(path string, timeout time.Duration) {
	err := WaitUntilExists(path, timeout)
	if err != nil {
		panic(err)
	}
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}

func IsDirectoryOrPanic(path string) bool {
	ok, err := IsDirectory(path)
	if err != nil {
		panic(err)
	}
	return ok
}

func Copy(f *os.File, w io.Writer, buffer []byte) (p int64, err error) {
	if f == nil {
		panic("cannot Copy nil file")
	}
	if w == nil {
		panic("cannot Copy into nil writer")
	}
	if buffer == nil {
		panic("cannot Copy into nil buffer")
	}
	n := -1
	for err != io.EOF && n != 0 {
		n, err = f.Read(buffer)
		if err == io.EOF || n == 0 {
			continue
		}
		if err != nil {
			return
		}
		n, err = w.Write(buffer[0:n])
		if err != nil {
			return
		}
		p += int64(n)
	}
	return p, nil
}

func CopyChunk(src *os.File, dest io.Writer, buf []byte, start, end int64) (int64, error) {
	if src == nil {
		panic("cannot Copy nil file")
	}
	if dest == nil {
		panic("cannot Copy into nil writer")
	}
	if buf == nil {
		panic("cannot Copy into nil buffer")
	}
	// By default limit scan to buffer size
	limit := len(buf)
	length := end - start
	if end > -1 && length < int64(limit) {
		// limit scan to (end - start) which is of type int
		limit = int(length)
	}

	// First src scan
	n, err := src.ReadAt(buf[0:limit], start)
	if err != nil && err != io.EOF {
		return 0, err
	}

	var total int64
	for n > 0 {
		// Loop while buffer is full
		k, err := dest.Write(buf[0:n])
		total += int64(k)
		if err != nil {
			return total, err
		}
		if k != n {
			err = fmt.Errorf("bytes count read and written mismatch")
			return total, err
		}

		// Adjust limit
		if end > -1 && (length-total) < int64(limit) {
			// Stop copy before EOF
			limit = int(length - total)
		}

		// Next src scan
		n, err = src.ReadAt(buf[0:limit], start+int64(total))
		if err != nil && err != io.EOF {
			return total, err
		}
	}

	return total, nil
}
