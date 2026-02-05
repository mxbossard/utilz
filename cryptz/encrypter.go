package cryptz

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"slices"

	"golang.org/x/crypto/chacha20poly1305"
)

type Encrypter interface {
	Encrypt(nonce []byte, plaintext []byte) ([]byte, error)
	Decrypt(nonce []byte, ciphertext []byte) ([]byte, error)
}

type AesGcmEncrypter struct {
	Encrypter
	kh keyHolder
}

func (e AesGcmEncrypter) Encrypt(nonce []byte, plaintext []byte) ([]byte, error) {
	key, err := e.kh.Key(32)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	encrypted := aesgcm.Seal(nil, nonce, plaintext, nil)

	return encrypted, nil
}

func (e AesGcmEncrypter) Decrypt(nonce []byte, ciphertext []byte) ([]byte, error) {
	key, err := e.kh.Key(32)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)

	return plaintext, err
}

type Chacha20poly1305Encrypter struct {
	Encrypter
	kh keyHolder
}

func (e Chacha20poly1305Encrypter) Encrypt(nonce []byte, plaintext []byte) ([]byte, error) {
	key, err := e.kh.Key(32)
	if err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	encrypted := aead.Seal(nil, nonce, plaintext, nil)

	return encrypted, nil
}

func (e Chacha20poly1305Encrypter) Decrypt(nonce []byte, ciphertext []byte) ([]byte, error) {
	key, err := e.kh.Key(32)
	if err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)

	return plaintext, err
}

type CompositeEncrypter struct {
	Encrypter
	encrypters []Encrypter
}

func (e CompositeEncrypter) Encrypt(nonce []byte, plaintext []byte) ([]byte, error) {
	ciphertext := plaintext
	var err error
	for _, encrypter := range e.encrypters {
		ciphertext, err = encrypter.Encrypt(nonce, ciphertext)
		if err != nil {
			return nil, err
		}
	}
	return ciphertext, nil
}

func (e CompositeEncrypter) Decrypt(nonce []byte, ciphertext []byte) ([]byte, error) {
	plaintext := ciphertext
	var err error
	reversed := e.encrypters
	slices.Reverse(reversed)
	for _, encrypter := range reversed {
		plaintext, err = encrypter.Decrypt(nonce, plaintext)
		if err != nil {
			return nil, err
		}
	}
	return plaintext, nil
}

type DummyEncrypter struct {
	Encrypter
	prefix, suffix string
}

func (e DummyEncrypter) Encrypt(nonce []byte, plaintext []byte) (r []byte, err error) {
	r = append([]byte(e.prefix), plaintext...)
	r = append(r, []byte(e.suffix)...)
	return
}

func (e DummyEncrypter) Decrypt(nonce []byte, ciphertext []byte) (r []byte, err error) {
	pLen := len([]byte(e.prefix))
	sLen := len([]byte(e.suffix))
	dLen := len(ciphertext) - pLen - sLen
	if dLen < 0 {
		err = fmt.Errorf("unable to decrypt data")
		return
	}
	r = ciphertext[pLen : pLen+dLen]
	return
}
