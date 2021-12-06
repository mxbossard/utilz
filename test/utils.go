package test

import (
	"fmt"
	"os"
	"path/filepath"
	//"github.com/stretchr/testify/assert"
	"time"
	"math/rand"
)

func init() {
	fmt.Println("Rand seed initialization ...")
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

func MkRandTempDir() (path string) {
        path, _ = BuildRandTempPath()
        os.MkdirAll(path, 0755)
        os.Chdir(path)
        return
}

