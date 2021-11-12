package lnurl

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

// HandleLNURL takes a bech32-encoded lnurl and either gets its parameters from the query-
// string or calls the URL to get the parameters.
// Returns a different struct for each of the lnurl subprotocols, the .LNURLKind() method of
// which should be checked next to see how the wallet is going to proceed.
func HandleLNURL(rawlnurl string) (string, LNURLParams, error) {
	var err error
	var rawurl string

	if strings.HasPrefix(rawlnurl, "https:") {
		rawurl = rawlnurl
	} else {
		lnurl, ok := FindLNURLInText(rawlnurl)
		if !ok {
			return "", nil, errors.New("invalid bech32-encoded lnurl: " + rawlnurl)
		}
		rawurl, err = LNURLDecode(lnurl)
		if err != nil {
			return "", nil, err
		}
	}

	parsed, err := url.Parse(rawurl)
	if err != nil {
		return rawurl, nil, err
	}

	query := parsed.Query()

	switch query.Get("tag") {
	case "login":
		value, err := HandleAuth(rawurl, parsed, query)
		return rawurl, value, err
	case "withdrawRequest":
		if value, ok := HandleFastWithdraw(query); ok {
			return rawurl, value, nil
		}
	}

	resp, err := http.Get(rawurl)
	if err != nil {
		return rawurl, nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rawurl, nil, err
	}

	j := gjson.ParseBytes(b)
	if j.Get("status").String() == "ERROR" {
		return rawurl, nil, LNURLErrorResponse{
			URL:    parsed,
			Reason: j.Get("reason").String(),
			Status: "ERROR",
		}
	}

	switch j.Get("tag").String() {
	case "withdrawRequest":
		value, err := HandleWithdraw(j)
		return rawurl, value, err
	case "payRequest":
		value, err := HandlePay(j)
		return rawurl, value, err
	case "channelRequest":
		value, err := HandleChannel(j)
		return rawurl, value, err
	default:
		return rawurl, nil, errors.New("unknown response tag " + j.String())
	}
}
