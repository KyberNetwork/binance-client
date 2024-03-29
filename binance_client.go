package binance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	defaultTimeout = 5 * time.Second
)

// Client to interact with binance api
type Client struct {
	httpClient       *http.Client
	apiKey           string
	secretKey        string
	apiBaseURL       string // for both spot and margin
	futureAPIBaseURL string
}

// NewClient create new client object
func NewClient(key, secret, apiBaseURL, futureAPIBaseURL string, hc *http.Client) *Client {
	return &Client{
		apiKey:           key,
		secretKey:        secret,
		apiBaseURL:       apiBaseURL,
		futureAPIBaseURL: futureAPIBaseURL,
		httpClient:       hc,
	}
}

func (bc *Client) createListenKey(apiPath string) (string, error) {
	var (
		listenKey ListenKey
	)
	requestURL := fmt.Sprintf("%s/%s", bc.apiBaseURL, apiPath)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return "", err
	}

	rr := req.WithHeader(apiKeyHeader, bc.apiKey).Request()
	_, err = bc.doRequest(rr, &listenKey)
	if err != nil {
		return "", err
	}
	return listenKey.ListenKey, nil
}

func (bc *Client) keepListenKeyAlive(listenKey, apiPath string) error {
	requestURL := fmt.Sprintf("%s/%s", bc.apiBaseURL, apiPath)
	req, err := NewRequestBuilder(http.MethodPut, requestURL, nil)
	if err != nil {
		return err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("listenKey", listenKey).
		Request()
	_, err = bc.doRequest(rr, nil)
	return err
}

func (bc *Client) doRequest(req *http.Request, data interface{}) (*FwdData, error) {
	resp, err := bc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute the request, %w", err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read response failed, %w", err)
	}
	fwd := &FwdData{
		Status:      resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Data:        respBody,
	}
	switch resp.StatusCode {
	case http.StatusOK:
		if data == nil { // if data == nil then caller does not care about response body, consider as success
			return fwd, nil
		}
		if err = json.Unmarshal(respBody, data); err != nil {
			return fwd, fmt.Errorf("failed to parse data into struct: %s %w", respBody, err)
		}
	default:
		var responseErr = struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}{}
		_ = json.Unmarshal(respBody, &responseErr)
		return fwd, fmt.Errorf("%w, raw: %d, %s: ", newAPIError(responseErr.Code, responseErr.Msg), resp.StatusCode, string(respBody))
	}
	return fwd, nil
}
