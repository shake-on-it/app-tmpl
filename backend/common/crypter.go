package common

import (
	"crypto/aes"
	"crypto/cipher"
	"io/ioutil"
)

type Crypter interface {
	Decrypt(val interface{}) (interface{}, error)
	Encrypt(val interface{}) (interface{}, error)
}

func LoadCrypter(path string) (Crypter, error) {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewCrypter(key)
}

func NewCrypter(key []byte) (Crypter, error) {
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &crypter{blockCipher}, nil
}

type crypter struct {
	blockCipher cipher.Block
}

func (c *crypter) Decrypt(val interface{}) (interface{}, error) {
	return val, nil
}

func (c *crypter) Encrypt(val interface{}) (interface{}, error) {
	return val, nil
}
