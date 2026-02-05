package cryptz

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

const (
	nonceSize = 12
)

// The most :
// A complete abstraction to os.File.
// Permit to write in the middle of the file.
// All writes are buffered in memory.
// A Flush encrypt all the data one shot.
// Appending to file
type EncryptedFile struct {
	*sync.Mutex
	*os.File

	encrypter Encrypter

	publicPath  string
	privatePath string
	buffer      []byte
	length      int64
	offset      int64
	readonly    bool
	append      bool
	closed      bool
	dirty       int64
}

// Close the file
func (f *EncryptedFile) Close() error {
	f.Lock()
	defer f.Unlock()
	if !f.readonly {
		err := f.sync()
		if err != nil {
			return err
		}
	}
	f.buffer = nil
	f.closed = true
	return f.File.Close()
}

// func (f *EncryptedFile) Name() string {
// 	return f.File.Name()
// }

func (f *EncryptedFile) ClearName() string {
	return filepath.Join(f.publicPath, f.privatePath)
}

func (f *EncryptedFile) Read(b []byte) (n int, err error) {
	f.Lock()
	defer f.Unlock()
	if f.closed {
		err = fmt.Errorf("file is closed")
		return
	}
	n = copy(b, f.buffer[f.offset:f.offset+f.length])
	return
}

func (f *EncryptedFile) ReadAt(b []byte, off int64) (n int, err error) {
	f.Lock()
	defer f.Unlock()
	if f.closed {
		err = fmt.Errorf("file is closed")
		return
	}
	n = copy(b, f.buffer[:f.length+off])
	return
}

func (f *EncryptedFile) ReadFrom(r io.Reader) (n int64, err error) {
	f.Lock()
	defer f.Unlock()
	if f.closed {
		err = fmt.Errorf("file is closed")
		return
	}
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
	f.Lock()
	defer f.Unlock()
	if f.closed {
		err = fmt.Errorf("file is closed")
		return
	}
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

// func (f *EncryptedFile) SetDeadline(t time.Time) error {
// 	return f.File.SetDeadline(t)
// }

// func (f *EncryptedFile) SetReadDeadline(t time.Time) error {
// 	return f.File.SetReadDeadline(t)
// }

// func (f *EncryptedFile) SetWriteDeadline(t time.Time) error {
// 	return f.File.SetWriteDeadline(t)
// }

// func (f *EncryptedFile) Stat() (os.FileInfo, error) {
// 	return f.File.Stat()
// }

func (f *EncryptedFile) Sync() error {
	f.Lock()
	defer f.Unlock()
	return f.sync()
}

func (f *EncryptedFile) SyscallConn() (syscall.RawConn, error) {
	return f.File.SyscallConn()
}

func (f *EncryptedFile) Truncate(size int64) error {
	f.Lock()
	defer f.Unlock()
	if f.closed {
		return fmt.Errorf("file is closed")
	}
	if f.readonly {
		err := fmt.Errorf("file is open readonly")
		return err
	}
	f.buffer = f.buffer[0:size]
	return nil
}

func (f *EncryptedFile) Write(b []byte) (n int, err error) {
	f.Lock()
	defer f.Unlock()
	if f.closed {
		err = fmt.Errorf("file is closed")
		return
	}
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
	if len(b) > 0 {
		f.dirty++
	}
	return len(b), nil
}

func (f *EncryptedFile) WriteAt(b []byte, off int64) (n int, err error) {
	f.Lock()
	defer f.Unlock()
	if f.closed {
		err = fmt.Errorf("file is closed")
		return
	}
	if f.readonly {
		err = fmt.Errorf("file is open readonly")
		return
	}
	end := f.buffer[off:]
	f.buffer = append(f.buffer[:off], b...)
	f.buffer = append(f.buffer, end...)
	if len(b) > 0 {
		f.dirty++
	}
	return len(f.buffer), nil
}

func (f *EncryptedFile) WriteString(s string) (n int, err error) {
	n, err = f.Write([]byte(s))
	if n > 0 {
		f.dirty++
	}
	return
}

func (f *EncryptedFile) WriteTo(w io.Writer) (n int64, err error) {
	f.Lock()
	defer f.Unlock()
	if f.closed {
		err = fmt.Errorf("file is closed")
		return
	}
	// FIXME: may need to manage an int64 file size with multiple successfis writes ?
	// fmt.Printf("writing %d bytes (%d) ...\n", len(f.buffer), f.length)
	l, err := w.Write(f.buffer[0:f.length])
	return int64(l), err
}

func (f *EncryptedFile) sync() error {
	if f.closed {
		return fmt.Errorf("file is closed")
	}
	if f.dirty == 0 {
		// nothing to do if not dirty
		return nil
	}
	nonce, err := ForgeNonce(12)
	if err != nil {
		return err
	}
	ciphertext, err := f.encrypter.Encrypt(nonce, f.buffer)
	if err != nil {
		return err
	}
	err = f.File.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.File.WriteAt(nonce, 0)
	if err != nil {
		return err
	}
	_, err = f.File.WriteAt(ciphertext, int64(len(nonce)))
	if err != nil {
		return err
	}
	err = f.File.Sync()
	f.dirty = 0
	return err
}

// Same as Sync
func (f *EncryptedFile) Flush() error {
	return f.Sync()
}

// Wipe all file updates, re-buffering file content in memory.
func (f *EncryptedFile) Resync() (err error) {
	f.Lock()
	defer f.Unlock()
	if f.closed {
		err = fmt.Errorf("file is closed")
		return
	}
	// FIXME: may need to manage an int64 file size with multiple successfis read ?
	buf := bytes.NewBuffer(make([]byte, 0, 1000))
	n, err := f.File.WriteTo(buf)
	if err != nil {
		return err
	}
	// fmt.Printf("copied %d bytes\n", n)
	if n > 0 {
		data := buf.Bytes()
		nonce := data[0:nonceSize]
		ciphertext := data[nonceSize:]
		plaintext, err := f.encrypter.Decrypt(nonce, ciphertext)
		if err != nil {
			return err
		}
		f.buffer = plaintext
		f.length = int64(len(plaintext))
	}
	f.dirty = 0
	// fmt.Printf("buffer: len=%d ; cap=%d bytes\n", len(f.buffer), cap(f.buffer))
	return nil
}
