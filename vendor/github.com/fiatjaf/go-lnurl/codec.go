package lnurl

import (
	"errors"
	"strings"
)

// LNURLDecode takes a bech32-encoded lnurl string and returns a plain-text https URL.
func LNURLDecode(lnurl string) (url string, err error) {
	tag, data, err := Decode(lnurl)
	if err != nil {
		return
	}

	if tag != "lnurl" {
		err = errors.New("tag is not 'lnurl', but '" + tag + "'")
		return
	}

	converted, err := ConvertBits(data, 5, 8, false)
	if err != nil {
		return
	}

	url = string(converted)
	return
}

// LNURLEncode takes a plain-text https URL and returns a bech32-encoded uppercased lnurl string.
func LNURLEncode(actualurl string) (lnurl string, err error) {
	asbytes := []byte(actualurl)
	converted, err := ConvertBits(asbytes, 8, 5, true)
	if err != nil {
		return
	}

	lnurl, err = Encode("lnurl", converted)
	return strings.ToUpper(lnurl), err
}
