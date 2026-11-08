// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package secret

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	_ "crypto/sha256" // registers crypto.SHA256
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// AesEncrypt encrypts text and given key with AES.
// It is only internally used at the moment to use "SECRET_KEY" for some database values.
func AesEncrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AesEncrypt invalid key: %v", err)
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("AesEncrypt unable to read IV: %w", err)
	}
	cfb := cipher.NewCFBEncrypter(block, iv) //nolint:staticcheck // need to migrate and refactor to a new approach
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

// AesDecrypt decrypts text and given key with AES.
// It is only internally used at the moment to use "SECRET_KEY" for some database values.
func AesDecrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("AesDecrypt ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv) //nolint:staticcheck // need to migrate and refactor to a new approach
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, fmt.Errorf("AesDecrypt invalid decrypted base64 string: %w", err)
	}
	return data, nil
}

// secretKeyHash turns SECRET_KEY into a 32-byte AES key: SHA-256, as every stored
// secret was encrypted with - a slower derivation would make them unreadable.
func secretKeyHash(key string) []byte {
	h := crypto.SHA256.New()
	h.Write([]byte(key))
	return h.Sum(nil)
}

// EncryptSecret encrypts a string with given key into a hex string
func EncryptSecret(key, str string) (string, error) {
	keyHash := secretKeyHash(key)
	plaintext := []byte(str)
	ciphertext, err := AesEncrypt(keyHash[:], plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt by secret: %w", err)
	}
	return hex.EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts a previously encrypted hex string
func DecryptSecret(key, cipherHex string) (string, error) {
	keyHash := secretKeyHash(key)
	ciphertext, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt by secret, invalid hex string: %w", err)
	}
	plaintext, err := AesDecrypt(keyHash[:], ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt by secret, the key (maybe SECRET_KEY?) might be incorrect: %w", err)
	}
	return string(plaintext), nil
}
