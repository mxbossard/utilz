package truzt

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/mod/sumdb/dirhash"

	"github.com/mxbossard/utilz/filez"
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

func SerializeFileInfo(f filez.File) ([]byte, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(nil)
	_, err = fmt.Fprintf(b, "%d-%d-%d-%s", info.ModTime(), info.Mode(), info.Size(), info.Name())
	return b.Bytes(), err
}

// Sign a file metadata
func SignFileInfo(f filez.File) (sign []byte, err error) {
	// FIXME: not optimized to scan multiple files
	hash, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	b, err := SerializeFileInfo(f)
	hash.Write(b)
	sign = hash.Sum(nil)
	return
}

// Sign a file including metadata
func SignFile(path string) (sign []byte, err error) {
	// FIXME: not optimized to scan multiple files
	sign1, err := SignFilesContent(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	sign2, err := SignFileInfo(f)
	if err != nil {
		return nil, err
	}
	hash, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	hash.Write(sign2)
	hash.Write([]byte(sign1))
	return hash.Sum(nil), nil
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
