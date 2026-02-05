package cryptz

import (
	"bytes"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

type keyHolder struct {
	passphrase []byte
	salt       []byte
}

func (h keyHolder) Key(length int) ([]byte, error) {
	hash := sha512.New
	info := "Key"
	salt, err := h.Salt(length)
	if err != nil {
		return nil, err
	}
	return hkdf.Key(hash, h.passphrase, salt, info, length)
}

func (h keyHolder) Salt(length int) ([]byte, error) {
	hash := sha512.New
	info := "Hash"
	return hkdf.Key(hash, h.salt, nil, info, length)
}

func (h keyHolder) SaltedHash(length int, data []byte) ([]byte, error) {
	hash := sha512.New()
	salt, err := h.Salt(length)
	if err != nil {
		return nil, err
	}
	hash.Write(data)
	hash.Write(salt)
	hashed := hash.Sum(nil)
	return hashed[0:length], nil
}

func NewKeyHolder(passphrase, salt []byte) keyHolder {
	return keyHolder{
		passphrase: passphrase,
		salt:       salt,
	}
}

func NewStringKeyHolder(passphrase, salt string) keyHolder {
	return NewKeyHolder([]byte(passphrase), []byte(salt))
}

func ForgeNonce(length int, datas ...any) ([]byte, error) {
	// Generate a nonce based on randomn, current timestamp, and supplied data
	buf := bytes.NewBuffer(nil)
	// Random
	rnd := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, rnd); err != nil {
		return nil, err
	}
	// Timestamp
	ns := time.Now().Nanosecond()
	binary.Write(buf, binary.LittleEndian, ns)
	//subtle.ConstantTimeCompare()

	for _, data := range datas {
		if d, ok := data.([]byte); ok {
			buf.Write(d)
		} else {
			err := fmt.Errorf("unable to cast data to byte slice")
			return nil, err
		}
	}

	hash := sha512.New
	info := "Nonce"
	return hkdf.Key(hash, rnd, buf.Bytes(), info, length)
}
