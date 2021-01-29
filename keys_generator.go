package main

import (
	"encoding/hex"
	log "log"

	"github.com/btcsuite/btcutil/base58"
	"github.com/muun/libwallet"
)

var defaultNetwork = libwallet.Mainnet()

func buildExtendedKeys(rawKey1, rawKey2, recoveryCode string) (
	*libwallet.DecryptedPrivateKey,
	*libwallet.DecryptedPrivateKey) {

	// Always take the salt from the second key (the same salt was used, but our older key format
	// is missing the salt on the first key):
	salt := extractSalt(rawKey2)

	decryptionKey, err := libwallet.RecoveryCodeToKey(recoveryCode, salt)
	if err != nil {
		log.Fatalf("failed to process recovery code: %v", err)
	}

	key1, err := decryptionKey.DecryptKey(rawKey1, defaultNetwork)
	if err != nil {
		log.Fatalf("failed to decrypt first key: %v", err)
	}

	key2, err := decryptionKey.DecryptKey(rawKey2, defaultNetwork)
	if err != nil {
		log.Fatalf("failed to decrypt second key: %v", err)
	}

	return key1, key2
}

func extractSalt(rawKey string) string {
	bytes := base58.Decode(rawKey)
	saltBytes := bytes[len(bytes)-8:]

	return hex.EncodeToString(saltBytes)
}
