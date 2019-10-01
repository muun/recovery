package main

import (
	log "log"

	"github.com/btcsuite/btcutil/base58"
	"github.com/muun/libwallet"
)

func buildExtendedKey(rawKey, recoveryCode string) *libwallet.DecryptedPrivateKey {
	recoveryCodeBytes := extractBytes(recoveryCode)
	salt := extractSalt(rawKey)

	privKey := libwallet.NewChallengePrivateKey(recoveryCodeBytes, salt)

	key, err := privKey.DecryptKey(rawKey, libwallet.Mainnet())
	if err != nil {
		log.Fatalf("failed to decrypt key: %v", err)
	}

	return key
}

func extractSalt(rawKey string) []byte {
	bytes := base58.Decode(rawKey)
	return bytes[len(bytes)-8:]
}

func extractBytes(recoveryCode string) []byte {
	return []byte(recoveryCode)
}
