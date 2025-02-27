package goribot

import (
	"bytes"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// PostDataType is the type of Content-Type
type PostDataType int

const (
	// TextPostData  text/plain
	TextPostData PostDataType = iota
	// UrlencodedPostData  application/x-www-form-urlencoded
	UrlencodedPostData
	// JsonPostData  application/json
	JsonPostData
)

// Request  is request struct
type Request struct {
	Url    *url.URL
	Method string
	Cookie []*http.Cookie
	Header http.Header
	Body   []byte
	Proxy  string
}

// SetHeader set Header of request
func (r *Request) SetHeader(k, v string) *Request {
	r.Header.Set(k, v)
	return r
}

// SetBody set body data of request
func (r *Request) SetBody(body []byte) *Request {
	r.Body = body
	return r
}

// AddCookie add a cookie to request
func (r *Request) AddCookie(k, v string) *Request {
	r.Cookie = append(r.Cookie, &http.Cookie{
		Name:  k,
		Value: v,
	})
	return r
}

// WithProxy set proxy of request
func (r *Request) WithProxy(proxy string) *Request {
	r.Proxy = proxy
	return r
}

// NewRequest create a new request
func NewRequest() *Request {
	return &Request{
		Url:    &url.URL{},
		Method: "GET",
		Cookie: []*http.Cookie{},
		Header: http.Header{},
		Body:   []byte{},
		Proxy:  "",
	}
}

// Response  is response struct
type Response struct {
	Url    *url.URL
	Status int
	Header http.Header
	Body   []byte

	Request      *Request
	HttpResponse *http.Response

	Text string
	Html *goquery.Document
	Json map[string]interface{}
}

// DefaultClient is the default Client and is used by Get, Head, and Post.
var DefaultClient = &http.Client{
	Timeout: 10 * time.Second,
}

// Download do a request return response and error
func Download(r *Request) (*Response, error) {
	HttpRequest, err := http.NewRequest(r.Method, r.Url.String(), bytes.NewReader(r.Body))
	if err != nil {
		return nil, err
	}

	if r.Header != nil {
		HttpRequest.Header = r.Header.Clone()
	}
	for _, i := range r.Cookie {
		HttpRequest.AddCookie(i)
	}

	var c *http.Client
	if r.Proxy != "" {
		c = &http.Client{
			Timeout: 10 * time.Second,
		}
		p, err := url.Parse(r.Proxy)
		if err != nil {
			return nil, err
		}
		c.Transport = &http.Transport{
			Proxy: http.ProxyURL(p),
		}
	} else {
		c = DefaultClient
	}

	resp, err := c.Do(HttpRequest)
	if err != nil {
		return nil, HttpErr{error: err, Request: r}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))
	var j map[string]interface{}
	_ = json.Unmarshal(body, &j)

	return &Response{
		Url:          r.Url,
		Status:       resp.StatusCode,
		Header:       resp.Header,
		Body:         body,
		Request:      r,
		HttpResponse: resp,
		Text:         string(body),
		Html:         html,
		Json:         j,
	}, nil
}

// HttpErr is type of downloader error
type HttpErr struct {
	error
	Request *Request
}

// MustParseUrl parse url from str,if get error will do panic
func MustParseUrl(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}
