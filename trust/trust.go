package trust

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/mod/sumdb/dirhash"
)

func signString(s string) (sign string, err error) {
	var hash = sha256.New()
	_, err = hash.Write([]byte(s))
	if err != nil {
		return "", err
	}
	ba := hash.Sum(nil)
	//sign = string(ba[:])
	sign = fmt.Sprintf("%x", ba)
	return
}

func SignStrings(ss ...string) (sign string, err error) {
	if len(ss) == 0 {
		return
	} else if len(ss) == 1 {
		return signString(ss[0])
	}

	concat := ""
	for _, s := range ss {
		sign, err = signString(s)
		if err != nil {
			return "", err
		}
		concat += sign + ";"
	}
	sign, err = signString(concat)
	return
}

func SignFilesContent(pathes ...string) (sign string, err error) {
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

func SignFsContents(pathes ...string) (sign string, err error) {
	signatures := map[string]string{}
	for _, path := range pathes {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if fileInfo.IsDir() {
			sign, err = SignDirContent(path)
		} else {
			sign, err = SignFilesContent(path)
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

func SignObject(object interface{}) (sign string, err error) {
	b, err := json.Marshal(object)
	if err != nil {
		return "", err
	}
	sign, err = SignStrings(string(b[:]))
	return sign, err
}

func SignObjects(objects ...interface{}) (sign string, err error) {
	concat := ""
	for _, object := range objects {
		s, err := SignObject(object)
		if err != nil {
			return "", err
		}
		concat += s + ";"
	}
	sign, err = SignStrings(concat)
	return
}
