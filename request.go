package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type RequestBuilder struct {
	req    *http.Request
	params url.Values
}

func NewRequestBuilder(method, url string, body io.ReadCloser) (*RequestBuilder, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return &RequestBuilder{
		req:    req,
		params: req.URL.Query(),
	}, nil
}

func (r *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	r.req.Header.Set(key, value)
	return r
}

func (r *RequestBuilder) WithParam(key, value string) *RequestBuilder {
	r.params.Set(key, value)
	return r
}

func (r *RequestBuilder) SignedRequest(secret string) *http.Request {
	r.params.Set("timestamp", strconv.FormatUint(getTimepoint(), 10))
	r.params.Set("recvWindow", "5000")
	sig := url.Values{}
	sig.Set("signature", sign(r.params.Encode(), secret))
	r.req.URL.RawQuery = r.params.Encode() + "&" + sig.Encode()
	return r.req
}

func (r *RequestBuilder) Request() *http.Request {
	return r.req
}

func sign(msg string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(msg)); err != nil {
		panic(err) // should never happen
	}
	result := ethereum.Bytes2Hex(mac.Sum(nil))
	return result
}

func getTimepoint() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
}
