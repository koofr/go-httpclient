package httpclient

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var XmlHeaderBytes []byte = []byte(xml.Header)

type HTTPClient struct {
	BaseURL          *url.URL
	Headers          http.Header
	Client           *http.Client
	PostHooks        map[int]func(*http.Request, *http.Response) error
	rateLimited      bool
	rateLimitChan    chan struct{}
	rateLimitTimeout time.Duration
}

func New() (httpClient *HTTPClient) {
	return &HTTPClient{
		Client:    HttpClient,
		Headers:   make(http.Header),
		PostHooks: make(map[int]func(*http.Request, *http.Response) error),
	}
}

func Insecure() (httpClient *HTTPClient) {
	httpClient = New()
	httpClient.Client = InsecureHttpClient
	return
}

func (c *HTTPClient) SetPostHook(onStatus int, hook func(*http.Request, *http.Response) error) {
	c.PostHooks[onStatus] = hook
}

func (c *HTTPClient) SetRateLimit(limit int, timeout time.Duration) {
	c.rateLimited = true
	c.rateLimitChan = make(chan struct{}, limit)

	for i := 0; i < limit; i++ {
		c.rateLimitChan <- struct{}{}
	}

	c.rateLimitTimeout = timeout
}

func (c *HTTPClient) buildURL(req *RequestData) *url.URL {
	bu := c.BaseURL

	opaque := EscapePath(bu.Path + req.Path)

	u := &url.URL{
		Scheme: bu.Scheme,
		Host:   bu.Host,
		Opaque: opaque,
	}

	if req.Params != nil {
		u.RawQuery = req.Params.Encode()
	}

	return u
}

func (c *HTTPClient) setHeaders(req *RequestData, httpReq *http.Request) {
	switch req.RespEncoding {
	case EncodingJSON:
		httpReq.Header.Set("Accept", "application/json")
	case EncodingXML:
		httpReq.Header.Set("Accept", "application/xml")
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
			lr := io.LimitReader(response.Body, 10*1024)
			contentBytes, _ := ioutil.ReadAll(lr)
			content := string(contentBytes)

			err = InvalidStatusError{req.ExpectedStatus, response.StatusCode, response.Header, content}
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

	case EncodingXML:
		defer response.Body.Close()

		if buf, err = ioutil.ReadAll(response.Body); err != nil {
			return
		}

		err = xml.Unmarshal(buf, req.RespValue)

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

	if req.RespConsume {
		defer response.Body.Close()
		ioutil.ReadAll(response.Body)
	}

	return
}

func (c *HTTPClient) marshalRequest(req *RequestData) (err error) {
	if req.ReqReader != nil || req.ReqValue == nil {
		return
	}

	if req.Headers == nil {
		req.Headers = make(http.Header)
	}

	var buf []byte

	switch req.ReqEncoding {
	case EncodingJSON:
		buf, err = json.Marshal(req.ReqValue)

		if err != nil {
			return
		}

		req.ReqReader = bytes.NewReader(buf)
		req.Headers.Set("Content-Type", "application/json")

		return

	case EncodingXML:
		buf, err = xml.Marshal(req.ReqValue)

		if err != nil {
			return
		}

		buf = append(XmlHeaderBytes, buf...)

		req.ReqReader = bytes.NewReader(buf)
		req.Headers.Set("Content-Type", "application/xml")

		return
	}

	err = fmt.Errorf("HTTPClient: invalid ReqEncoding: %s", req.ReqEncoding)

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
	err = c.marshalRequest(req)

	if err != nil {
		return
	}

	r, err := http.NewRequest(req.Method, req.FullURL, req.ReqReader)

	if err != nil {
		return
	}

	if req.FullURL == "" {
		r.URL = c.buildURL(req)
		r.Host = r.URL.Host
	}

	c.setHeaders(req, r)

	if c.rateLimited {
		if c.rateLimitTimeout > 0 {
			select {
			case t := <-c.rateLimitChan:
				defer func() {
					c.rateLimitChan <- t
				}()
			case <-time.After(c.rateLimitTimeout):
				return nil, RateLimitTimeoutError
			}
		} else {
			t := <-c.rateLimitChan
			defer func() {
				c.rateLimitChan <- t
			}()
		}
	}

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
		defer response.Body.Close()
		return
	}

	if err = c.unmarshalResponse(req, response); err != nil {
		return
	}

	return
}
