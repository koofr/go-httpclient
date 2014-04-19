package httpclient

import (
	"io"
	"net/http"
	"net/url"
)

type Encoding string

const (
	EncodingJSON = "JSON"
	EncodingXML  = "XML"
)

type RequestData struct {
	Method          string
	Path            string
	Params          url.Values
	FullURL         string // client.BaseURL + Path or FullURL
	Headers         http.Header
	ReqReader       io.Reader
	ReqEncoding     Encoding
	ReqValue        interface{}
	ExpectedStatus  []int
	IgnoreRedirects bool
	RespEncoding    Encoding
	RespValue       interface{}
	RespConsume     bool
}
