package lnurl

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

var lnurlregex = regexp.MustCompile(`.*?((lnurl)([0-9]{1,}[a-z0-9]+){1})`)

// FindLNURLInText uses a Regular Expression to find a bech32-encoded lnurl string in a blob of text.
func FindLNURLInText(text string) (lnurl string, ok bool) {
	text = strings.ToLower(text)
	results := lnurlregex.FindStringSubmatch(text)

	if len(results) == 0 {
		return
	}

	return results[1], true
}

// RandomK1 returns a 32-byte random hex-encoded string for usage as k1 in lnurl-auth and anywhere else.
func RandomK1() string {
	random := make([]byte, 32)
	rand.Read(random)
	return hex.EncodeToString(random)
}
