package libwallet

import (
	hash "golang.org/x/crypto/ripemd160"
)

func ripemd160(data []byte) []byte {
	hasher := hash.New()
	_, err := hasher.Write(data)
	if err != nil {
		panic("failed to hash")
	}

	return hasher.Sum([]byte{})
}
