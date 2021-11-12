package lnurl

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

var (
	f     bool  = false
	t     bool  = true
	FALSE *bool = &f
	TRUE  *bool = &t
)

func Action(text string, url string) *SuccessAction {
	if url == "" {
		return &SuccessAction{
			Tag:     "message",
			Message: text,
		}
	}

	if text == "" {
		text = " "
	}
	return &SuccessAction{
		Tag:         "url",
		Description: text,
		URL:         url,
	}
}

func AESAction(description string, preimage []byte, content string) (*SuccessAction, error) {
	plaintext := []byte(content)

	ciphertext, iv, err := AESCipher(preimage, plaintext)
	if err != nil {
		return nil, err
	}

	return &SuccessAction{
		Tag:         "aes",
		Description: description,
		Ciphertext:  base64.StdEncoding.EncodeToString(ciphertext),
		IV:          base64.StdEncoding.EncodeToString(iv),
	}, nil
}

type LNURLPayResponse1 struct {
	LNURLResponse
	Callback        string   `json:"callback"`
	CallbackURL     *url.URL `json:"-"`
	Tag             string   `json:"tag"`
	MaxSendable     int64    `json:"maxSendable"`
	MinSendable     int64    `json:"minSendable"`
	EncodedMetadata string   `json:"metadata"`
	Metadata        Metadata `json:"-"`
	CommentAllowed  int64    `json:"commentAllowed"`
}

type LNURLPayResponse2 struct {
	LNURLResponse
	SuccessAction *SuccessAction `json:"successAction"`
	Routes        [][]RouteInfo  `json:"routes"`
	PR            string         `json:"pr"`
	Disposable    *bool          `json:"disposable,omitempty"`
}

type RouteInfo struct {
	NodeId        string `json:"nodeId"`
	ChannelUpdate string `json:"channelUpdate"`
}

type SuccessAction struct {
	Tag         string `json:"tag"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Message     string `json:"message,omitempty"`
	Ciphertext  string `json:"ciphertext,omitempty"`
	IV          string `json:"iv,omitempty"`
}

func (sa *SuccessAction) Decipher(preimage []byte) (content string, err error) {
	ciphertext, err := base64.StdEncoding.DecodeString(sa.Ciphertext)
	if err != nil {
		return
	}

	iv, err := base64.StdEncoding.DecodeString(sa.IV)
	if err != nil {
		return
	}

	plaintext, err := AESDecipher(preimage, ciphertext, iv)
	if err != nil {
		return
	}

	return string(plaintext), nil
}

func (_ LNURLPayResponse1) LNURLKind() string { return "lnurl-pay" }

func HandlePay(j gjson.Result) (LNURLParams, error) {
	strmetadata := j.Get("metadata").String()
	var metadata Metadata
	err := json.Unmarshal([]byte(strmetadata), &metadata)
	if err != nil {
		return nil, err
	}

	callback := j.Get("callback").String()

	// parse url
	callbackURL, err := url.Parse(callback)
	if err != nil {
		return nil, errors.New("callback is not a valid URL")
	}

	// add random nonce to avoid caches
	qs := callbackURL.Query()
	qs.Set("nonce", strconv.FormatInt(time.Now().Unix(), 10))
	callbackURL.RawQuery = qs.Encode()

	return LNURLPayResponse1{
		Tag:             "payRequest",
		Callback:        callback,
		CallbackURL:     callbackURL,
		EncodedMetadata: strmetadata,
		Metadata:        metadata,
		MaxSendable:     j.Get("maxSendable").Int(),
		MinSendable:     j.Get("minSendable").Int(),
		CommentAllowed:  j.Get("commentAllowed").Int(),
	}, nil
}

type Metadata [][]string

// Description returns the content of text/plain metadata entry.
func (m Metadata) Description() string {
	for _, entry := range m {
		if len(entry) == 2 && entry[0] == "text/plain" {
			return entry[1]
		}
	}
	return ""
}

// ImageDataURI returns image in the form data:image/type;base64,... if an image exists
// or an empty string if not.
func (m Metadata) ImageDataURI() string {
	for _, entry := range m {
		if len(entry) == 2 && strings.Split(entry[0], "/")[0] == "image" {
			return "data:" + entry[0] + "," + entry[1]
		}
	}
	return ""
}

// ImageBytes returns image as bytes, decoded from base64 if an image exists
// or nil if not.
func (m Metadata) ImageBytes() []byte {
	for _, entry := range m {
		if len(entry) == 2 && strings.Split(entry[0], "/")[0] == "image" {
			if decoded, err := base64.StdEncoding.DecodeString(entry[1]); err == nil {
				return decoded
			}
		}
	}
	return nil
}

// ImageExtension returns the file extension for the image, either "png" or "jpeg"
func (m Metadata) ImageExtension() string {
	for _, entry := range m {
		if len(entry) == 2 && strings.Split(entry[0], "/")[0] == "image" {
			spl := strings.Split(entry[0], "/")
			if len(spl) == 2 {
				return strings.Split(spl[1], ";")[0]
			}
		}
	}
	return ""
}

// Entry returns an arbitrary entry from the metadata array.
// eg.: "video/mp4" or "application/vnd.some-specific-thing-from-a-specific-app".
func (m Metadata) Entry(key string) string {
	for _, entry := range m {
		if len(entry) == 2 && entry[0] == key {
			return entry[1]
		}
	}
	return ""
}
