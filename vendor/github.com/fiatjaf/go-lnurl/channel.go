package lnurl

import (
	"errors"
	"net/url"

	"github.com/tidwall/gjson"
)

type LNURLChannelResponse struct {
	LNURLResponse
	Tag         string   `json:"tag"`
	K1          string   `json:"k1"`
	Callback    string   `json:"callback"`
	CallbackURL *url.URL `json:"-"`
	URI         string   `json:"uri"`
}

func (_ LNURLChannelResponse) LNURLKind() string { return "lnurl-channel" }

func HandleChannel(j gjson.Result) (LNURLParams, error) {
	k1 := j.Get("k1").String()
	if k1 == "" {
		return nil, errors.New("k1 is blank")
	}
	callback := j.Get("callback").String()
	callbackURL, err := url.Parse(callback)
	if err != nil {
		return nil, errors.New("callback is not a valid URL")
	}

	return LNURLChannelResponse{
		Tag:         "channelRequest",
		K1:          k1,
		Callback:    callback,
		CallbackURL: callbackURL,
		URI:         j.Get("uri").String(),
	}, nil
}
