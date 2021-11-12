package lnurl

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/tidwall/gjson"
)

type LNURLWithdrawResponse struct {
	LNURLResponse
	Tag                string   `json:"tag"`
	K1                 string   `json:"k1"`
	Callback           string   `json:"callback"`
	CallbackURL        *url.URL `json:"-"`
	MaxWithdrawable    int64    `json:"maxWithdrawable"`
	MinWithdrawable    int64    `json:"minWithdrawable"`
	DefaultDescription string   `json:"defaultDescription"`
	BalanceCheck       string   `json:"balanceCheck,omitempty"`
}

func (_ LNURLWithdrawResponse) LNURLKind() string { return "lnurl-withdraw" }

func HandleWithdraw(j gjson.Result) (LNURLParams, error) {
	callback := j.Get("callback").String()
	callbackURL, err := url.Parse(callback)
	if err != nil {
		return nil, errors.New("callback is not a valid URL")
	}

	return LNURLWithdrawResponse{
		Tag:                "withdrawRequest",
		K1:                 j.Get("k1").String(),
		Callback:           callback,
		CallbackURL:        callbackURL,
		MaxWithdrawable:    j.Get("maxWithdrawable").Int(),
		MinWithdrawable:    j.Get("minWithdrawable").Int(),
		DefaultDescription: j.Get("defaultDescription").String(),
		BalanceCheck:       j.Get("balanceCheck").String(),
	}, nil
}

func HandleFastWithdraw(query url.Values) (LNURLParams, bool) {
	callback := query.Get("callback")
	if callback == "" {
		return nil, false
	}
	callbackURL, err := url.Parse(callback)
	if err != nil {
		return nil, false
	}
	maxWithdrawable, err := strconv.ParseInt(query.Get("maxWithdrawable"), 10, 64)
	if err != nil {
		return nil, false
	}
	minWithdrawable, err := strconv.ParseInt(query.Get("minWithdrawable"), 10, 64)
	if err != nil {
		return nil, false
	}
	balanceCheck := query.Get("balanceCheck")

	return LNURLWithdrawResponse{
		Tag:                "withdrawRequest",
		K1:                 query.Get("k1"),
		Callback:           callback,
		CallbackURL:        callbackURL,
		MaxWithdrawable:    maxWithdrawable,
		MinWithdrawable:    minWithdrawable,
		DefaultDescription: query.Get("defaultDescription"),
		BalanceCheck:       balanceCheck,
	}, true
}
