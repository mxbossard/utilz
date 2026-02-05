package cryptz

import (
	"crypto/sha512"
	"encoding/base64"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/blake2b"

	"github.com/mxbossard/utilz/filez"
)

type Namer interface {
	// Hashed filepath is composed of a public & private path
	// public path is not hashed then private path is hached
	HashFilepath(publicPath, privatePath string) (string, error)
}

type Sha512Blake2Namer struct {
	Namer
	kh keyHolder
}

func (n Sha512Blake2Namer) HashFilepath(publicPath, privatePath string) (string, error) {
	salt, err := n.kh.Salt(96)
	if err != nil {
		return "", err
	}
	hasher1 := sha512.New()
	hasher2, err := blake2b.New512(nil)
	if err != nil {
		return "", err
	}
	path := publicPath
	privatePathes := filez.SplitPath(privatePath)

	sw := strings.Builder{}
	encoder := base64.NewEncoder(base64.URLEncoding, &sw)
	for _, p := range privatePathes {
		p += string(salt)
		hasher1.Reset()
		hasher2.Reset()
		hasher1.Write([]byte(p))
		hashed := hasher1.Sum(nil)
		hasher2.Write(hashed)
		hashed = hasher2.Sum(nil)
		// FIXME: convert hashed into a valid path string
		sw.Reset()
		encoder.Write(hashed)
		encodedPath := sw.String()
		truncatedPath := encodedPath[0:min(len(encodedPath), 200)]
		path = filepath.Join(path, truncatedPath)
	}
	err = encoder.Close()

	return path, err
}

type ClearTextNamer struct {
	Namer
	prefix, suffix string
}

func (n ClearTextNamer) HashFilepath(publicPath, privatePath string) (string, error) {
	return filepath.Join(publicPath, privatePath), nil
}

type DummyNamer struct {
	Namer
	prefix, suffix string
}

func (n DummyNamer) HashFilepath(publicPath, privatePath string) (string, error) {
	path := publicPath
	privatePathes := filez.SplitPath(privatePath)
	for _, p := range privatePathes {
		path = filepath.Join(path, n.prefix+p+n.suffix)
	}
	return path, nil
}
