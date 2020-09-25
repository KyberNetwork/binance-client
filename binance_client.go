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
}

// FwdData contain data we forward to client
type FwdData struct {
	Status      int
	ContentType string
	Data        []byte
}

// NewBinanceClient create new client object
func NewBinanceClient(key, secret string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		apiKey:     key,
		secretKey:  secret,
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
	)
	requestURL := fmt.Sprintf("%s/api/v3/userDataStream", apiBaseURL)
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

func (bc *Client) doRequest(req *http.Request, data interface{}) (*FwdData, error) {
	resp, err := bc.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute the request")
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response failed")
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
			return fwd, errors.Wrap(err, "failed to parse data into struct")
		}
	default:
		return fwd, errors.Errorf("%d, %s", resp.StatusCode, string(respBody))
	}
	return fwd, nil
}

// KeepListenKeyAlive keep it alive
func (bc *Client) KeepListenKeyAlive(listenKey string) error {
	requestURL := fmt.Sprintf("%s/api/v3/userDataStream", apiBaseURL)
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

// GetAccountState return account info
func (bc *Client) GetAccountState() (AccountState, error) {
	var (
		response AccountState
	)
	requestURL := fmt.Sprintf("%s/api/v3/account", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return AccountState{}, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).SignedRequest(bc.secretKey)
	_, err = bc.doRequest(rr, &response)
	return response, err
}

// CreateOrder create a limit order
func (bc *Client) CreateOrder(side, symbol, ordType, timeInForce, price, amount string) (CreateOrderResult, *FwdData, error) {
	var (
		response CreateOrderResult
	)
	requestURL := fmt.Sprintf("%s/api/v3/order", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return response, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("side", side).
		WithParam("type", ordType).
		WithParam("timeInForce", timeInForce).
		WithParam("quantity", amount).
		WithParam("price", price).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &response)
	return response, fwd, err
}

// GetOpenOrders return account info, if symbol is empty, all open order will return
func (bc *Client) GetOpenOrders(symbol string) ([]*OpenOrder, *FwdData, error) {
	var (
		response = make([]*OpenOrder, 0)
	)
	requestURL := fmt.Sprintf("%s/api/v3/openOrders", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey)
	if symbol != "" {
		rr = rr.WithParam("symbol", symbol)
	}
	rq := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rq, &response)
	return response, fwd, err
}
func (bc *Client) OrderStatus(symbol string, id int64) (*OpenOrder, *FwdData, error) {
	result := OpenOrder{}
	requestURL := fmt.Sprintf("%s/api/v3/order", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("orderId", strconv.FormatInt(id, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	return &result, fwd, err
}

// GetTradeHistory query recent trade list
func (bc *Client) GetTradeHistory(symbol string, limit int64) (TradeHistoryList, *FwdData, error) {
	result := TradeHistoryList{}
	requestURL := fmt.Sprintf("%s/api/v3/trades", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, nil, err
	}
	rr := req.
		WithParam("symbol", symbol).
		WithParam("limit", strconv.FormatInt(limit, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	return result, fwd, err
}

// GetAccountTradeHistory query account recent trade list
func (bc *Client) GetAccountTradeHistory(base, quote string, limit int64, fromID string) (AccountTradeHistoryList, *FwdData, error) {
	result := AccountTradeHistoryList{}
	symbol := strings.ToUpper(fmt.Sprintf("%s%s", base, quote))
	requestURL := fmt.Sprintf("%s/api/v3/myTrades", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
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
	fwd, err := bc.doRequest(signedReq, &result)
	return result, fwd, err
}

// WithdrawHistory query recent withdraw list
func (bc *Client) WithdrawHistory(fromMillis, toMillis int64) (WithdrawalsList, *FwdData, error) {
	result := WithdrawalsList{}
	requestURL := fmt.Sprintf("%s/wapi/v3/withdrawHistory.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return WithdrawalsList{}, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("startTime", strconv.FormatInt(fromMillis, 10)).
		WithParam("endTime", strconv.FormatInt(toMillis, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	if !result.Success && fwd != nil {
		return result, fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result, fwd, err
}

// DepositHistory query recent withdraw list
func (bc *Client) DepositHistory(fromMillis, toMillis int64) (DepositsList, *FwdData, error) {
	result := DepositsList{}
	requestURL := fmt.Sprintf("%s/wapi/v3/depositHistory.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return DepositsList{}, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("startTime", strconv.FormatInt(fromMillis, 10)).
		WithParam("endTime", strconv.FormatInt(toMillis, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	if !result.Success && fwd != nil {
		return result, fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result, fwd, err
}

// CancelOrder cancel an order
func (bc *Client) CancelOrder(symbol string, id int64) (CancelResult, *FwdData, error) {
	result := CancelResult{}
	requestURL := fmt.Sprintf("%s/api/v3/order", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodDelete, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("orderId", strconv.FormatInt(id, 10)).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	return result, fwd, err
}

// CancelAllOrder cancel all orders
func (bc *Client) CancelAllOrder(symbol string) ([]BOrder, *FwdData, error) {
	var result []BOrder
	requestURL := fmt.Sprintf("%s/api/v3/openOrders", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodDelete, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	return result, fwd, err
}

// Withdraw ...
func (bc *Client) Withdraw(symbol string, amount string, address string) (string, *FwdData, error) {
	var result WithdrawResult
	requestURL := fmt.Sprintf("%s/wapi/v3/withdraw.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return "", nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("address", address).
		WithParam("name", "reserve").
		WithParam("amount", amount).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return "", fwd, err
	}
	if !result.Success && fwd != nil {
		return "", fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result.ID, fwd, err
}

// WithdrawToMainAccount withdraw from sub account to main account
func (bc *Client) WithdrawToMainAccount(asset, amount string) (string, *FwdData, error) {
	var (
		result TransferToMasterResponse
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/sub-account/transfer/subToMaster", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return txID, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("asset", asset).
		WithParam("amount", amount).
		SignedRequest(bc.secretKey)
	fws, err := bc.doRequest(rr, &result)
	return result.TxID, err
}

// GetDepositAddress ...
func (bc *Client) GetDepositAddress(asset string) (BDepositAddress, *FwdData, error) {
	var result BDepositAddress
	requestURL := fmt.Sprintf("%s/wapi/v3/depositAddress.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("asset", asset).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	if !result.Success && fwd != nil {
		return result, fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result, fwd, err
}

// GetAllAssetDetail ...
func (bc *Client) GetAllAssetDetail(asset string) (AssetDetailResult, *FwdData, error) {
	var result AssetDetailResult
	requestURL := fmt.Sprintf("%s/wapi/v3/assetDetail.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	if !result.Success && fwd != nil {
		return result, fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result, fwd, err
}

// GetExchangeInfo ...
func (bc *Client) GetExchangeInfo() (ExchangeInfo, *FwdData, error) {
	var result ExchangeInfo
	requestURL := fmt.Sprintf("%s/api/v3/exchangeInfo", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.Request()
	fwd, err := bc.doRequest(rr, &result)

	return result, fwd, err
}

// GetServerTime ...
func (bc *Client) GetServerTime() (int64, *FwdData, error) {
	var result ServerTime
	requestURL := fmt.Sprintf("%s/api/v3/time", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return 0, nil, err
	}
	rr := req.Request()
	fwd, err := bc.doRequest(rr, &result)

	return result.ServerTime, fwd, err
}
