package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Encoding string

const (
	EncodingJSON = "JSON"
)

type RequestData struct {
	Method          string
	Path            string
	Params          url.Values
	FullURL         string // optional
	Headers         http.Header
	ReqReader       io.Reader
	ReqEncoding     Encoding
	ReqValue        interface{}
	ExpectedStatus  []int
	IgnoreRedirects bool
	RespEncoding    Encoding
	RespValue       interface{}
}

type InvalidStatusError struct {
	Expected []int
	Got      int
}

func (e InvalidStatusError) Error() string {
	return fmt.Sprintf("Invalid response status! Got %d, expected %d", e.Got, e.Expected)
}

type HTTPClient struct {
	BaseURL   *url.URL
	Headers   http.Header
	Client    *http.Client
	PostHooks map[int]func(*http.Request, *http.Response) error
}

func New() (httpClient *HTTPClient) {
	return &HTTPClient{
		Client:    HttpClient,
		Headers:   make(http.Header),
		PostHooks: make(map[int]func(*http.Request, *http.Response) error),
	}
}

func Insecure() (httpClient *HTTPClient) {
	return &HTTPClient{
		Client:    InsecureHttpClient,
		Headers:   make(http.Header),
		PostHooks: make(map[int]func(*http.Request, *http.Response) error),
	}
}

func (c *HTTPClient) SetPostHook(onStatus int, hook func(*http.Request, *http.Response) error) {
	c.PostHooks[onStatus] = hook
}

func (c *HTTPClient) buildURL(req *RequestData) string {
	if req.FullURL != "" {
		return req.FullURL
	}

	bu := c.BaseURL

	u := url.URL{
		Scheme: bu.Scheme,
		Host:   bu.Host,
		Path:   bu.Path + req.Path,
	}

	if req.Params != nil {
		u.RawQuery = req.Params.Encode()
	}

	return u.String()
}

func (c *HTTPClient) setHeaders(req *RequestData, httpReq *http.Request) {

	switch req.ReqEncoding {
	case EncodingJSON:
		httpReq.Header.Set("Content-Type", "application/json")
	}

	switch req.RespEncoding {
	case EncodingJSON:
		httpReq.Header.Set("Accept", "application/json")
	}

	if c.Headers != nil {
		for key, values := range c.Headers {
			for _, value := range values {
				httpReq.Header.Set(key, value)
			}
		}
	}

	if req.Headers != nil {
		for key, values := range req.Headers {
			for _, value := range values {
				httpReq.Header.Set(key, value)
			}
		}
	}
}

func (c *HTTPClient) checkStatus(req *RequestData, response *http.Response) (err error) {
	if req.ExpectedStatus != nil {
		statusOk := false

		for _, status := range req.ExpectedStatus {
			if response.StatusCode == status {
				statusOk = true
			}
		}

		if !statusOk {
			err = InvalidStatusError{req.ExpectedStatus, response.StatusCode}
			return
		}
	}

	return
}

func (c *HTTPClient) unmarshalResponse(req *RequestData, response *http.Response) (err error) {
	var buf []byte

	switch req.RespEncoding {
	case EncodingJSON:
		defer response.Body.Close()

		if buf, err = ioutil.ReadAll(response.Body); err != nil {
			return
		}

		err = json.Unmarshal(buf, req.RespValue)

		return
	}

	switch req.RespValue.(type) {
	case *[]byte:
		defer response.Body.Close()

		if buf, err = ioutil.ReadAll(response.Body); err != nil {
			return
		}

		respVal := req.RespValue.(*[]byte)
		*respVal = buf

		return
	}

	return
}

func (c *HTTPClient) marshalRequest(req *RequestData) (err error) {
	if req.ReqValue != nil && req.ReqEncoding != "" && req.ReqReader == nil {
		var buf []byte
		buf, err = json.Marshal(req.ReqValue)
		if err != nil {
			return
		}
		req.ReqReader = bytes.NewReader(buf)
	}
	return
}

func (c *HTTPClient) runPostHook(req *http.Request, response *http.Response) (err error) {
	hook, ok := c.PostHooks[response.StatusCode]

	if ok {
		err = hook(req, response)
	}

	return
}

func (c *HTTPClient) Request(req *RequestData) (response *http.Response, err error) {
	reqURL := c.buildURL(req)

	err = c.marshalRequest(req)

	if err != nil {
		return
	}

	r, err := http.NewRequest(req.Method, reqURL, req.ReqReader)

	if err != nil {
		return
	}

	c.setHeaders(req, r)

	if req.IgnoreRedirects {
		transport := c.Client.Transport

		if transport == nil {
			transport = http.DefaultTransport
		}

		response, err = transport.RoundTrip(r)
	} else {
		response, err = c.Client.Do(r)
	}

	if err != nil {
		return
	}

	if err = c.runPostHook(r, response); err != nil {
		return
	}

	if err = c.checkStatus(req, response); err != nil {
		return
	}

	if err = c.unmarshalResponse(req, response); err != nil {
		return
	}

	return
}