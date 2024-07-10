package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

type PasswordCrypter interface {
	Encrypt(text string) (string, error)
	Decrypt(text string) (string, error)
}

type Service struct {
	secret string
}

func NewPasswordCrypter(secret string) *Service {
	return &Service{secret: secret}
}

func (h *Service) createHash() []byte {
	hash := sha256.Sum256([]byte(h.secret))
	return hash[:]
}

func (h *Service) Encrypt(text string) (string, error) {
	block, err := aes.NewCipher(h.createHash())
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)
	return base64.URLEncoding.EncodeToString(cipherText), nil
}

func (h *Service) Decrypt(text string) (string, error) {
	cipherText, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(h.createHash())
	if err != nil {
		return "", err
	}
	if len(cipherText) < aes.BlockSize {
		return "", errors.New("cipher text too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)
	return string(cipherText), nil
}
