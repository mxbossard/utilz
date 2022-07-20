package trust

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"

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

func SignFsContents(pathes []string) (sign string, err error) {
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

func SignString(s string) (sign string, err error) {
	var hash = sha256.New()
	_, err = hash.Write([]byte(s))
	if err != nil {
		return "", err
	}
	ba := hash.Sum(nil)
	sign = string(ba[:])
	return
}

/*
func sortArray(objects []interface{}) error {
	if len(objects) == 0 {
		return
	}
	switch o := objects[0].(type) {
	case string:
		sort.Strings(objects.([]string))
	default:
		return fmt.Errorf("Not supported type: %T", o)
	}
}
*/

func SignObject(object interface{}) (sign string, err error) {
	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.String:
		sign, err = SignString(v.String())

	case reflect.Slice:
		concat := ""
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			s, err := SignObject(item)
			if err != nil {
				return "", err
			}
			concat += s + ";"
		}
		return SignObject(concat)

	case reflect.Map:
		concat := ""
		//sortedKeys := sortArray(v.MapKeys())
		sortedKeys := v.MapKeys()
		for _, key := range sortedKeys {
			val := v.MapIndex(key)
			s1, err := SignObject(key)
			if err != nil {
				return "", err
			}
			s2, err := SignObject(val)
			if err != nil {
				return "", err
			}
			concat += s1 + ":" + s2 + ";"
		}
		return SignObject(concat)

	default:
		err = fmt.Errorf("Not support object type: %T", object)
	}
	return
}
