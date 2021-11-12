package lnurl

import (
	"net/url"
)

// The base response for all lnurl calls.
type LNURLResponse struct {
	Status string `json:"status,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type LNURLErrorResponse struct {
	Status string   `json:"status,omitempty"`
	Reason string   `json:"reason,omitempty"`
	URL    *url.URL `json:"-"`
}

func (r LNURLErrorResponse) Error() string {
	return r.Reason
}

func OkResponse() LNURLResponse {
	return LNURLResponse{Status: "OK"}
}

func ErrorResponse(reason string) LNURLErrorResponse {
	return LNURLErrorResponse{
		URL:    nil,
		Status: "ERROR",
		Reason: reason,
	}
}

type LNURLParams interface {
	LNURLKind() string
}
