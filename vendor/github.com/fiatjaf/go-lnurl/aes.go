package lnurl

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func AESCipher(key, plaintext []byte) (ciphertext []byte, iv []byte, err error) {
	pad := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padding := make([]byte, pad)
	for i := 0; i < pad; i++ {
		padding[i] = byte(pad)
	}
	plaintext = append(plaintext, padding...)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	ciphertext = make([]byte, len(plaintext))
	iv = make([]byte, aes.BlockSize)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(ciphertext, plaintext)
	return
}

func AESDecipher(key, ciphertext, iv []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	size := len(ciphertext)
	pad := ciphertext[size-1]
	plaintext = ciphertext[:size-int(pad)]

	return plaintext, nil
}
