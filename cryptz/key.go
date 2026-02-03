package cryptz

import (
	"crypto/hkdf"
	"crypto/sha512"
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

func NewKeyHolder(passphrase, salt []byte) keyHolder {
	return keyHolder{
		passphrase: passphrase,
		salt:       salt,
	}
}

func NewStringKeyHolder(passphrase, salt string) keyHolder {
	return NewKeyHolder([]byte(passphrase), []byte(salt))
}
