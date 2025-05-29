package main

import (
	"fmt"

	"github.com/muun/libwallet"
	"github.com/muun/libwallet/emergencykit"
)

var defaultNetwork = libwallet.Mainnet()

func decodeKeysFromInput(rawKey1 string, rawKey2 string) ([]*libwallet.EncryptedPrivateKeyInfo, error) {
	key1, err := libwallet.DecodeEncryptedPrivateKey(rawKey1)
	if err != nil {
		return nil, fmt.Errorf("failed to decode first key: %w", err)
	}

	key2, err := libwallet.DecodeEncryptedPrivateKey(rawKey2)
	if err != nil {
		return nil, fmt.Errorf("failed to decode second key: %w", err)
	}

	return []*libwallet.EncryptedPrivateKeyInfo{key1, key2}, nil
}

func decodeKeysFromMetadata(meta *emergencykit.Metadata) ([]*libwallet.EncryptedPrivateKeyInfo, error) {
	decodedKeys := make([]*libwallet.EncryptedPrivateKeyInfo, len(meta.EncryptedKeys))

	for i, metaKey := range meta.EncryptedKeys {
		decodedKeys[i] = &libwallet.EncryptedPrivateKeyInfo{
			Version:      meta.Version,
			Birthday:     meta.BirthdayBlock,
			EphPublicKey: metaKey.DhPubKey,
			CipherText:   metaKey.EncryptedPrivKey,
			Salt:         metaKey.Salt,
		}
	}

	return decodedKeys, nil
}

func decryptKeys(encryptedKeys []*libwallet.EncryptedPrivateKeyInfo, recoveryCode string) ([]*libwallet.DecryptedPrivateKey, error) {
	// Always take the salt from the second key (the same salt was used for all keys, but our legacy
	// key format did not include it in the first key):
	salt := encryptedKeys[1].Salt

	decryptionKey, err := libwallet.RecoveryCodeToKey(recoveryCode, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to process recovery code: %w", err)
	}

	decryptedKeys := make([]*libwallet.DecryptedPrivateKey, len(encryptedKeys))

	for i, encryptedKey := range encryptedKeys {
		decryptedKey, err := decryptionKey.DecryptKey(encryptedKey, defaultNetwork)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt key %d: %w", i, err)
		}

		decryptedKeys[i] = decryptedKey
	}

	return decryptedKeys, nil
}
