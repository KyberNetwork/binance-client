package binance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/KyberNetwork/cex_account_data/common"
	"github.com/KyberNetwork/cex_account_data/lib/caller"
)

const (
	defaultTimeout = 5 * time.Second
	apiBaseURL     = "https://api.binance.com"
	apiKeyHeader   = "X-MBX-APIKEY"
)

// Client to interact with binance api
type Client struct {
	httpClient *http.Client
	apiKey     string
	secretKey  string
	sugar      *zap.SugaredLogger
}

// FwdData contain data we forward to client
type FwdData struct {
	Status      int
	ContentType string
	Data        []byte
}

// NewBinanceClient create new client object
func NewBinanceClient(key, secret string, sugar *zap.SugaredLogger) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		apiKey:     key,
		secretKey:  secret,
		sugar:      sugar,
	}
}

// ListenKey is listen for user data stream
type ListenKey struct {
	ListenKey string `json:"listenKey"`
}

// CreateListenKey create a listen key for user data stream
func (bc *Client) CreateListenKey() (string, error) {
	var (
		listenKey ListenKey
		logger    = bc.sugar.With("func", caller.GetCurrentFunctionName())
	)
	requestURL := fmt.Sprintf("%s/api/v3/userDataStream", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		logger.Errorw("failed to create request to create listen key", "error", err)
		return "", err
	}

	rr := req.WithHeader(apiKeyHeader, bc.apiKey).Request()
	_, err = bc.doRequest(rr, logger, &listenKey)
	if err != nil {
		return "", err
	}
	return listenKey.ListenKey, nil
}

func (bc *Client) doRequest(req *http.Request, logger *zap.SugaredLogger, data interface{}) (*FwdData, error) {
	resp, err := bc.httpClient.Do(req)
	if err != nil {
		logger.Errorw("failed to execute the request", "error", err)
		return nil, errors.Wrap(err, "failed to execute the request")
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorw("failed to read response body", "error", err)
		return nil, err
	}
	_ = resp.Body.Close()
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
			logger.Errorw("failed to parse data into struct", "error", err)
			return fwd, errors.Wrap(err, "failed to parse data into struct")
		}
	default:
		logger.Errorw("got unexpected status code", "code", resp.StatusCode, "responseBody", string(respBody))
		return fwd, fmt.Errorf("got unexpected status code %d, body=%s", resp.StatusCode, string(respBody))
	}
	return fwd, nil
}

// KeepListenKeyAlive keep it alive
func (bc *Client) KeepListenKeyAlive(listenKey string) error {
	var (
		logger = bc.sugar.With("func", caller.GetCurrentFunctionName())
	)
	requestURL := fmt.Sprintf("%s/api/v3/userDataStream", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPut, requestURL, nil)
	if err != nil {
		logger.Errorw("failed to create new request for keep listen key alive", "error", err)
		return err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("listenKey", listenKey).
		Request()
	_, err = bc.doRequest(rr, logger, nil)
	return err
}

// GetAccountState return account info
func (bc *Client) GetAccountState() (common.AccountState, error) {
	var (
		logger   = bc.sugar.With("func", caller.GetCurrentFunctionName())
		response common.AccountState
	)
	requestURL := fmt.Sprintf("%s/api/v3/account", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		logger.Errorw("failed to create new request for get account info", "error", err)
		return common.AccountState{}, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).SignedRequest(bc.secretKey)
	_, err = bc.doRequest(rr, logger, &response)
	return response, err
}

// GetOpenOrders return account info
func (bc *Client) GetOpenOrders() ([]*common.OpenOrder, error) {
	var (
		logger   = bc.sugar.With("func", caller.GetCurrentFunctionName())
		response = make([]*common.OpenOrder, 0)
	)
	requestURL := fmt.Sprintf("%s/api/v3/openOrders", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		logger.Errorw("failed to create new request for get account info", "error", err)
		return nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).SignedRequest(bc.secretKey)
	_, err = bc.doRequest(rr, logger, &response)
	return response, err
}
func (bc *Client) OrderStatus(symbol string, id int64) (*common.OpenOrder, *FwdData, error) {
	result := common.OpenOrder{}
	requestURL := fmt.Sprintf("%s/api/v3/order", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for get account info", "error", err)
		return nil, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("orderId", strconv.FormatInt(id, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, bc.sugar, &result)

	return &result, fwd, err
}
