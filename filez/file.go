package filez

import (
	"fmt"
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
