package main

import (
	log "log"

	"github.com/btcsuite/btcutil/base58"
	"github.com/muun/libwallet"
)

func buildExtendedKey(rawKey, recoveryCode string) *libwallet.DecryptedPrivateKey {
	salt := extractSalt(rawKey)

	decryptionKey, err := libwallet.RecoveryCodeToKey(recoveryCode, salt)
	if err != nil {
		log.Fatalf("failed to process recovery code: %v", err)
	}

	walletKey, err := decryptionKey.DecryptKey(rawKey, libwallet.Mainnet())
	if err != nil {
		log.Fatalf("failed to decrypt key: %v", err)
	}

	return walletKey
}

func extractSalt(rawKey string) string {
	bytes := base58.Decode(rawKey)
	saltBytes := bytes[len(bytes)-8:]

	return string(saltBytes)
}
