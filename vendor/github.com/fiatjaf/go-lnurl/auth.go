package lnurl

import (
	"encoding/hex"
	"errors"
	"net/url"

	"github.com/btcsuite/btcd/btcec"
)

type LNURLAuthParams struct {
	Tag      string
	K1       string
	Callback string
	Host     string
}

func (_ LNURLAuthParams) LNURLKind() string { return "lnurl-auth" }

// VerifySignature takes the hex-encoded parameters passed to an lnurl-login endpoint and verifies
// the signature against the key and challenge.
func VerifySignature(k1, sig, key string) (ok bool, err error) {
	bk1, err1 := hex.DecodeString(k1)
	bsig, err2 := hex.DecodeString(sig)
	bkey, err3 := hex.DecodeString(key)
	if err1 != nil || err2 != nil || err3 != nil {
		return false, errors.New("Failed to decode hex.")
	}

	pubkey, err := btcec.ParsePubKey(bkey, btcec.S256())
	if err != nil {
		return false, errors.New("Failed to parse pubkey: " + err.Error())
	}

	signature, err := btcec.ParseDERSignature(bsig, btcec.S256())
	if err != nil {
		return false, errors.New("Failed to parse signature: " + err.Error())
	}

	return signature.Verify(bk1, pubkey), nil
}

func HandleAuth(rawurl string, parsed *url.URL, query url.Values) (LNURLParams, error) {
	k1 := query.Get("k1")
	if _, err := hex.DecodeString(k1); err != nil || len(k1) != 64 {
		return nil, errors.New("k1 is not a valid 32-byte hex-encoded string.")
	}

	return LNURLAuthParams{
		Tag:      "login",
		K1:       k1,
		Callback: rawurl,
		Host:     parsed.Host,
	}, nil
}
