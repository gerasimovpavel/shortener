package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gerasimovpavel/shortener.git/internal/config"
)

func Encrypt(src string) (string, error) {
	key := sha256.Sum256([]byte(config.Options.PassphraseKey))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	enc := aesgcm.Seal(nil, nonce, []byte(src), nil)
	dst := hex.EncodeToString(enc)
	return dst, nil
}

func Decrypt(src string) (string, error) {
	key := sha256.Sum256([]byte(config.Options.PassphraseKey))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	encrypted, err := hex.DecodeString(src)
	if err != nil {
		return "", err
	}
	decrypted, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
