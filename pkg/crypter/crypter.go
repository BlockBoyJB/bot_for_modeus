package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

type Crypter interface {
	Encrypt(text string) (string, error)
	Decrypt(text string) (string, error)
}

const defaultSaltBlockSize = 16

type crypter struct {
	secret string
}

func NewCrypter(secret string) Crypter {
	return &crypter{secret: secret}
}

func (c *crypter) createKey(salt []byte) []byte {
	return pbkdf2.Key([]byte(c.secret), salt, 100000, 32, sha256.New)
}

func (c *crypter) Encrypt(text string) (string, error) {
	salt := make([]byte, defaultSaltBlockSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	block, err := aes.NewCipher(c.createKey(salt))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(text), nil)
	return base64.URLEncoding.EncodeToString(append(salt, cipherText...)), nil
}

func (c *crypter) Decrypt(text string) (string, error) {
	cipherText, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	if len(cipherText) < defaultSaltBlockSize {
		return "", errors.New("cipher text too short")
	}

	salt, cipherText := cipherText[:defaultSaltBlockSize], cipherText[defaultSaltBlockSize:]

	block, err := aes.NewCipher(c.createKey(salt))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aesGCM.NonceSize() {
		return "", errors.New("cipher text too short")
	}

	nonce, cipherText := cipherText[:aesGCM.NonceSize()], cipherText[aesGCM.NonceSize():]
	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
