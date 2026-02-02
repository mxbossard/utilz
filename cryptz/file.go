package cryptz

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Namer interface {
	// Hashed filepath is composed of a public & private path
	// public path is not hashed then private path is hached
	HashFilepath(publicPath, privatePath string) (string, error)
}

type Encrypter interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

type BasicNamer struct {
	Namer
}

func (n BasicNamer) HashFilepath(publicPath, privatePath string) (string, error) {
	panic("not implemented yet")
}

type BasicEncrypter struct {
	Encrypter
	key []byte
}

func (e BasicEncrypter) Encrypt(data []byte) ([]byte, error) {
	panic("not implemented yet")
}

func (e BasicEncrypter) Decrypt(data []byte) ([]byte, error) {
	panic("not implemented yet")
}

// The most :
// A complete abstraction to os.File.
// Permit to write in the middle of the file.
// All writes are buffered in memory.
// A Flush encrypt all the data one shot.
// Appending to file
type EncryptedFile struct {
	encrypted *os.File

	// key       []byte
	namer     Namer
	encrypter Encrypter

	publicPath  string
	privatePath string
	buffer      []byte
	length      int64
	offset      int64
	readonly    bool
	append      bool
	flushed     bool
}

func (f *EncryptedFile) Close() error {
	f.buffer = nil
	return f.encrypted.Close()
}

func (f *EncryptedFile) Name() string {
	return filepath.Join(f.publicPath, f.privatePath)
}

func (f *EncryptedFile) Read(b []byte) (n int, err error) {
	n = copy(b, f.buffer[f.offset:f.offset+f.length])
	return
}

func (f *EncryptedFile) ReadAt(b []byte, off int64) (n int, err error) {
	n = copy(b, f.buffer[:f.length+off])
	return
}

func (f *EncryptedFile) ReadFrom(r io.Reader) (n int64, err error) {
	if f.readonly {
		err = fmt.Errorf("file is open readonly")
		return
	}
	buf := bytes.NewBuffer(nil)
	n, err = buf.ReadFrom(r)
	f.length = n
	f.buffer = buf.Bytes()
	return
}

func (f *EncryptedFile) Seek(offset int64, whence int) (ret int64, err error) {
	var newOffset int64
	switch whence {
	case 0:
		// relative to the origin of the file
		newOffset = offset
	case 1:
		// relative to the current offset
		newOffset += offset
	case 2:
		// relative to the current offset
		newOffset = f.length + offset
	default:
		err = fmt.Errorf("unknown whence: %d", whence)
	}

	if newOffset < 0 {
		err = fmt.Errorf("offset cannot be negative")
	} else if newOffset < f.length {
		err = fmt.Errorf("offset cannot be greater than file length")
	}
	if err != nil {
		return
	}

	ret = newOffset
	f.offset = newOffset
	return
}

// func (f *EncryptedFile) SetDeadline(t time.Time) error
// func (f *EncryptedFile) SetReadDeadline(t time.Time) error
// func (f *EncryptedFile) SetWriteDeadline(t time.Time) error
// func (f *EncryptedFile) Stat() (os.FileInfo, error)

func (f *EncryptedFile) Sync() error {
	e, err := f.encrypter.Encrypt(f.buffer)
	if err != nil {
		return err
	}
	err = f.encrypted.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.WriteAt(e, 0)
	if err != nil {
		return err
	}
	return nil
}

// func (f *EncryptedFile) SyscallConn() (syscall.RawConn, error)

func (f *EncryptedFile) Truncate(size int64) error {
	if f.readonly {
		err := fmt.Errorf("file is open readonly")
		return err
	}
	f.buffer = f.buffer[0:size]
	return nil
}

func (f *EncryptedFile) Write(b []byte) (n int, err error) {
	if f.readonly {
		err = fmt.Errorf("file is open readonly")
		return
	}
	if f.append {
		// Appending mode
		f.buffer = append(f.buffer, b...)
	} else {
		f.buffer = append(f.buffer[:f.offset], b...)
	}
	return len(b), nil
}

func (f *EncryptedFile) WriteAt(b []byte, off int64) (n int, err error) {
	if f.readonly {
		err = fmt.Errorf("file is open readonly")
		return
	}
	end := f.buffer[off:]
	f.buffer = append(f.buffer[:off], b...)
	f.buffer = append(f.buffer, end...)
	return len(f.buffer), nil
}

func (f *EncryptedFile) WriteString(s string) (n int, err error) {
	if f.readonly {
		err = fmt.Errorf("file is open readonly")
		return
	}
	return f.Write([]byte(s))
}

func (f *EncryptedFile) WriteTo(w io.Writer) (n int64, err error) {
	if f.readonly {
		err = fmt.Errorf("file is open readonly")
		return
	}
	// FIXME: may need to manage an int64 file size ?
	l, err := w.Write(f.buffer)
	return int64(l), err
}

func (f *EncryptedFile) Resync() (err error) {
	// TODO read encrypted data and buffer it
	buf := bytes.NewBuffer(nil)
	n, err := f.encrypted.ReadFrom(buf)
	if err != nil {
		return err
	}
	f.buffer = buf.Bytes()
	f.length = n
	return nil
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
func Open(key []byte, publicPath, privatePath string) (*EncryptedFile, error) {
	return OpenFile(key, publicPath, privatePath, os.O_RDONLY, 0)
}

// OpenFile is the generalized open call; most users will use Open
// or Create instead. It opens the named file with specified flag
// (O_RDONLY etc.). If the file does not exist, and the O_CREATE flag
// is passed, it is created with mode perm (before umask). If successful,
// methods on the returned File can be used for I/O.
func OpenFile(key []byte, publicPath, privatePath string, flag int, perms os.FileMode) (*EncryptedFile, error) {
	panic("not implemented yet")
	namer := &BasicNamer{}
	encrypter := &BasicEncrypter{key: key}
	f := &EncryptedFile{
		namer:       namer,
		encrypter:   encrypter,
		publicPath:  publicPath,
		privatePath: privatePath,
		readonly:    true,
	}
	name, err := namer.HashFilepath(publicPath, privatePath)
	if err != nil {
		return nil, err
	}
	f.encrypted, err = os.OpenFile(name, flag, perms)
	if err != nil {
		return nil, err
	}
	err = f.Resync()
	return f, err
}

func Write(key []byte, data []byte, publicPath, privatePath string) error {
	panic("not implemented yet")
}

func WriteString(key []byte, data, publicPath, privatePath string) error {
	panic("not implemented yet")
}

func Read(key []byte, publicPath, privatePath string) ([]byte, error) {
	panic("not implemented yet")
}

func ReadString(key []byte, publicPath, privatePath string) (string, error) {
	panic("not implemented yet")
}
