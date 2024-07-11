package filez

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func manageError(err error) bool {
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return false
	}
	return true
}

// Create a directory. It may already exists.
func CreateDirectory(path string) (err error) {
	err = os.MkdirAll(path, 0755)
	return
}

// Create a new directory.
func CreateNewDirectory(path string) (err error) {
	err = os.Mkdir(path, 0755)
	return
}

// Create a directory in a parent directory.
func CreateSubDirectory(parentDirPath, name string) (path string, err error) {
	path = filepath.Join(parentDirPath, name)
	err = CreateDirectory(path)
	return
}

// Create a new directory in a parent directory.
func CreateNewSubDirectory(parentDirPath, name string) (path string, err error) {
	path = filepath.Join(parentDirPath, name)
	err = CreateNewDirectory(path)
	return
}

// Get working directory path.
// Fail if cannot get working directory.
func WorkDirPath() (path string, err error) {
	path, err = os.Getwd()
	return
}

func Chdir(path string) (err error) {
	err = os.Chdir(path)
	return
}

// Create a file with a string content only if file does not exists yet.
func SoftInitFile(filepath, content string) (path string, err error) {
	_, err = os.Stat(filepath)
	if os.IsNotExist(err) {
		// Do not overwrite file if it already exists
		err = os.WriteFile(filepath, []byte(content), 0644)
	}
	return
}

func Read(filepath string) (content []byte, err error) {
	content, err = os.ReadFile(filepath)
	return
}

func ReadString(filepath string) (content string, err error) {
	var bytes []byte
	bytes, err = os.ReadFile(filepath)
	if err != nil {
		return
	}
	content = string(bytes)
	return
}

func WriteString(filepath, content string, perm os.FileMode) (err error) {
	bytes := []byte(content)
	err = os.WriteFile(filepath, bytes, perm)
	return
}

func Print(filepath string) (err error) {
	content, err := Read(filepath)
	if err != nil {
		return
	}
	fmt.Printf("\n%s\n", content)
	return
}

func PrintTree(parentPath string) {
	filepath.Walk(parentPath, func(name string, info os.FileInfo, err error) error {
		fmt.Println(name)
		return nil
	})
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}

func MkTemp(pattern string) string {
	dir := os.TempDir()
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, f.Name())
	return path
}

func MkTemp2(dir, pattern string) string {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, f.Name())
	return path
}

func MkTempDir(dir, pattern string) string {
	p, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		panic(err)
	}
	return p
}

func Open(dir, pattern string) *os.File {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		panic(err)
	}
	return f
}

func ReadFile(f *os.File, size int) []byte {
	b := make([]byte, size)
	n, err := f.ReadAt(b, 0)
	if err != nil {
		panic(err)
	}
	return b[:n]
}

func ReadFileString(f *os.File, size int) string {
	return string(ReadFile(f, size))
}

func Copy(f *os.File, w io.Writer, buffer []byte) (p int64, err error) {
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
