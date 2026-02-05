package cryptz

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/mxbossard/utilz/filez"
)

type encryptionFactory struct {
	namer     Namer
	encrypter Encrypter
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
func (f encryptionFactory) Open(key []byte, publicPath, privatePath string) (*EncryptedFile, error) {
	return f.OpenFile(key, publicPath, privatePath, os.O_RDONLY, 0400)
}

// OpenFile is the generalized open call; most users will use Open
// or Create instead. It opens the named file with specified flag
// (O_RDONLY etc.). If the file does not exist, and the O_CREATE flag
// is passed, it is created with mode perm (before umask). If successful,
// methods on the returned File can be used for I/O.
func (f encryptionFactory) OpenFile(key []byte, publicPath, privatePath string, flag int, perms os.FileMode) (*EncryptedFile, error) {
	ef := &EncryptedFile{
		Mutex:       &sync.Mutex{},
		encrypter:   f.encrypter,
		publicPath:  publicPath,
		privatePath: privatePath,
		readonly:    flag&0x3 == os.O_RDONLY,
	}
	name, err := f.namer.HashFilepath(publicPath, privatePath)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(name)
	err = filez.MkdirAll(dir, filez.DefaultDirPerms)
	if err != nil {
		return nil, err
	}

	ef.File, err = os.OpenFile(name, flag, perms)
	if err != nil {
		return nil, err
	}
	err = ef.Resync()
	return ef, err
}

func (f encryptionFactory) Write(key []byte, data []byte, publicPath, privatePath string) error {
	panic("not implemented yet")
}

func (f encryptionFactory) WriteString(key []byte, data, publicPath, privatePath string) error {
	panic("not implemented yet")
}

func (f encryptionFactory) Read(key []byte, publicPath, privatePath string) ([]byte, error) {
	panic("not implemented yet")
}

func (f encryptionFactory) ReadString(key []byte, publicPath, privatePath string) (string, error) {
	panic("not implemented yet")
}

func NewDefaultNamer(kh keyHolder) Namer {
	return &Sha512Blake2Namer{kh: kh}
}

func NewCleartextNamer(kh keyHolder) Namer {
	return &Sha512Blake2Namer{kh: kh}
}

func NewDefaultEncrypter(kh keyHolder) Encrypter {
	return &CompositeEncrypter{
		encrypters: []Encrypter{
			AesGcmEncrypter{
				kh: kh,
			},
			Chacha20poly1305Encrypter{
				kh: kh,
			},
		},
	}
}

func NewEncryptionFactory(namer Namer, encrypter Encrypter) encryptionFactory {
	return encryptionFactory{
		namer:     namer,
		encrypter: encrypter,
	}
}

func NewDefaultEncryptionFactory(kh keyHolder) encryptionFactory {
	return NewEncryptionFactory(NewDefaultNamer(kh), NewDefaultEncrypter(kh))
}
