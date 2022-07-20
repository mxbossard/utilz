package test

import (
	//"fmt"
	"os"
	"path/filepath"

	//"github.com/stretchr/testify/assert"
	"math/rand"
	"time"
)

func init() {
	//fmt.Println("Rand seed initialization ...")
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func BuildRandTempPath() (filePath string, err error) {
	fileName := RandSeq(10)
	filePath = filepath.Join(os.TempDir(), fileName)
	err = os.RemoveAll(filePath)
	if err != nil {
		return
	}

	return
}

func MkRandTempFile() (path string, err error) {
	path, err = BuildRandTempPath()
	if err != nil {
		return
	}
	err = os.WriteFile(path, []byte(""), 0400)
	return
}

func MkRandTempDir() (path string, err error) {
	path, err = BuildRandTempPath()
	if err != nil {
		return
	}
	err = os.MkdirAll(path, 0755)
	//os.Chdir(path)
	return
}

func MkRandSubDir(parentPath string) (path string, err error) {
	fileName := RandSeq(6)
	path = filepath.Join(parentPath, fileName)
	err = os.RemoveAll(path)
	if err != nil {
		return
	}
	err = os.MkdirAll(path, 0755)
	//os.Chdir(path)
	return
}
