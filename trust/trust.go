package trust

import (
	"crypto/sha256"
	"errors"
	"io"
	"os"

	"golang.org/x/mod/sumdb/dirhash"
)

func SignFileContent(path string) (sign string, err error) {
	return SignFilesContent([]string{path})
}

func SignFilesContent(pathes []string) (sign string, err error) {
	open := func(filePath string) (io.ReadCloser, error) {
		return os.Open(filePath)
	}
	sign, err = dirhash.Hash1(pathes, open)
	if err != nil {
		return
	}
	return
}

func SignDirContent(path string) (sign string, err error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !fileInfo.IsDir() {
		return "", errors.New("Supplied path is not a directory")
	}

	sign, err = dirhash.HashDir(path, "", dirhash.Hash1)
	if err != nil {
		return
	}
	return
}

func SignContents(pathes []string) (sign string, err error) {
	signatures := map[string]string{}
	for _, path := range pathes {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if fileInfo.IsDir() {
			sign, err = SignDirContent(path)
		} else {
			sign, err = SignFileContent(path)
		}
		if err != nil {
			return "", err
		}
		signatures[path] = sign
	}

	var hash = sha256.New()
	for _, path := range pathes {
		h, ok := signatures[path]
		if !ok {
			continue
		}
		msg := path + ":" + h + ";"
		_, err = hash.Write([]byte(msg))
		//fmt.Printf("Added hash: %s\n", msg)
		if err != nil {
			return "", err
		}
	}

	ba := hash.Sum(nil)
	sign = string(ba[:])
	return
}
