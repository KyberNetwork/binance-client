package binance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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
		return fwd, errors.Errorf("%d, %s", resp.StatusCode, string(respBody))
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

// GetOpenOrders return account info, if symbol is empty, all open order will return
func (bc *Client) GetOpenOrders(symbol string) ([]*common.OpenOrder, *FwdData, error) {
	var (
		logger   = bc.sugar.With("func", caller.GetCurrentFunctionName())
		response = make([]*common.OpenOrder, 0)
	)
	requestURL := fmt.Sprintf("%s/api/v3/openOrders", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		logger.Errorw("failed to create new request for open orders", "error", err)
		return nil, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey)
	if symbol != "" {
		rr = rr.WithParam("symbol", symbol)
	}
	rq := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rq, logger, &response)
	return response, fwd, err
}
func (bc *Client) OrderStatus(symbol string, id int64) (*common.OpenOrder, *FwdData, error) {
	result := common.OpenOrder{}
	requestURL := fmt.Sprintf("%s/api/v3/order", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for order status", "error", err)
		return nil, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("orderId", strconv.FormatInt(id, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, bc.sugar, &result)
	return &result, fwd, err
}

// GetTradeHistory query recent trade list
func (bc *Client) GetTradeHistory(symbol string, limit int64) (common.BinanceTradeHistory, *FwdData, error) {
	result := common.BinanceTradeHistory{}
	requestURL := fmt.Sprintf("%s/api/v3/trades", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for get trade history", "error", err)
		return nil, nil, err
	}
	rr := req.
		WithParam("symbol", symbol).
		WithParam("limit", strconv.FormatInt(limit, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, bc.sugar, &result)
	return result, fwd, err
}

// GetAccountTradeHistory query account recent trade list
func (bc *Client) GetAccountTradeHistory(base, quote string, limit int64, fromID string) (common.BinanceAccountTradeHistory, *FwdData, error) {
	result := common.BinanceAccountTradeHistory{}
	symbol := strings.ToUpper(fmt.Sprintf("%s%s", base, quote))
	requestURL := fmt.Sprintf("%s/api/v3/myTrades", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for get account trade history", "error", err)
		return nil, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("limit", strconv.FormatInt(limit, 10))
	if fromID != "" {
		rr = rr.WithParam("fromId", fromID)
	} else {
		rr = rr.WithParam("fromId", "0")
	}
	signedReq := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(signedReq, bc.sugar, &result)
	return result, fwd, err
}

// WithdrawHistory query recent withdraw list
func (bc *Client) WithdrawHistory(fromMillis, toMillis int64) (common.BinanceWithdrawals, *FwdData, error) {
	result := common.BinanceWithdrawals{}
	requestURL := fmt.Sprintf("%s/wapi/v3/withdrawHistory.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for get withdraw history", "error", err)
		return common.BinanceWithdrawals{}, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("startTime", strconv.FormatInt(fromMillis, 10)).
		WithParam("endTime", strconv.FormatInt(toMillis, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, bc.sugar, &result)
	if err != nil {
		return result, fwd, err
	}
	if !result.Success && fwd != nil {
		return result, fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result, fwd, err
}

// DepositHistory query recent withdraw list
func (bc *Client) DepositHistory(fromMillis, toMillis int64) (common.BinanceDeposits, *FwdData, error) {
	result := common.BinanceDeposits{}
	requestURL := fmt.Sprintf("%s/wapi/v3/depositHistory.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for get deposit history", "error", err)
		return common.BinanceDeposits{}, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("startTime", strconv.FormatInt(fromMillis, 10)).
		WithParam("endTime", strconv.FormatInt(toMillis, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, bc.sugar, &result)
	if err != nil {
		return result, fwd, err
	}
	if !result.Success && fwd != nil {
		return result, fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result, fwd, err
}

// CancelOrder cancel an order
func (bc *Client) CancelOrder(symbol string, id int64) (common.BinanceCancel, *FwdData, error) {
	result := common.BinanceCancel{}
	requestURL := fmt.Sprintf("%s/api/v3/order", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodDelete, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for cancel order", "error", err)
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("orderId", strconv.FormatInt(id, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, bc.sugar, &result)
	return result, fwd, err
}

// CancelAllOrder cancel all orders
func (bc *Client) CancelAllOrder(symbol string) ([]common.BinanceOrder, *FwdData, error) {
	var result []common.BinanceOrder
	requestURL := fmt.Sprintf("%s/api/v3/openOrders", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodDelete, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for get cancel all orders", "error", err)
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, bc.sugar, &result)
	return result, fwd, err
}

// Withdraw ...
func (bc *Client) Withdraw(symbol string, amount string, address string) (string, *FwdData, error) {
	var result common.BinanceWithdrawResult
	requestURL := fmt.Sprintf("%s/wapi/v3/withdraw.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for withdraw", "error", err)
		return "", nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("address", address).
		WithParam("name", "reserve").
		WithParam("amount", amount).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, bc.sugar, &result)
	if err != nil {
		return "", fwd, err
	}
	if !result.Success && fwd != nil {
		return "", fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result.ID, fwd, err
}

// GetDepositAddress ...
func (bc *Client) GetDepositAddress(asset string) (common.BinanceDepositAddress, *FwdData, error) {
	var result common.BinanceDepositAddress
	requestURL := fmt.Sprintf("%s/wapi/v3/depositAddress.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for get deposit address", "error", err)
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("asset", asset).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, bc.sugar, &result)
	if err != nil {
		return result, fwd, err
	}
	if !result.Success && fwd != nil {
		return result, fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result, fwd, err
}

// GetAllAssetDetail ...
func (bc *Client) GetAllAssetDetail(asset string) (common.BinanceAssetDetailResult, *FwdData, error) {
	var result common.BinanceAssetDetailResult
	requestURL := fmt.Sprintf("%s/wapi/v3/assetDetail.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for get asset detail", "error", err)
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, bc.sugar, &result)
	if err != nil {
		return result, fwd, err
	}
	if !result.Success && fwd != nil {
		return result, fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result, fwd, err
}

// GetExchangeInfo ...
func (bc *Client) GetExchangeInfo() (common.BinanceExchangeInfo, *FwdData, error) {
	var result common.BinanceExchangeInfo
	requestURL := fmt.Sprintf("%s/api/v3/exchangeInfo", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for get exchange info", "error", err)
		return result, nil, err
	}
	rr := req.Request()
	fwd, err := bc.doRequest(rr, bc.sugar, &result)

	return result, fwd, err
}

// getServerTime ...
func (bc *Client) getServerTime() (int64, *FwdData, error) {
	var result common.ServerTime
	requestURL := fmt.Sprintf("%s/api/v3/time", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		bc.sugar.Errorw("failed to create new request for get exchange info", "error", err)
		return 0, nil, err
	}
	rr := req.Request()
	fwd, err := bc.doRequest(rr, bc.sugar, &result)

	return result.ServerTime, fwd, err
}
