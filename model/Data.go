package model

import (
	"io"
	"net/http"
	"net/url"
)

type Data struct {
	Address string
	URL     *url.URL
	Path    string
	Query   url.Values
	Method  string
	Header  http.Header
	Body    io.ReadCloser
}
