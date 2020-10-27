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
			return fwd, errors.Wrapf(err, "failed to parse data into struct: %s", respBody)
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
func (bc *Client) CreateOrder(side, symbol, ordType, timeInForce, price, quantity string) (CreateOrderResult, *FwdData, error) {
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
		WithParam("quantity", quantity).
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

// OrderStatus ...
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
func (bc *Client) WithdrawHistory(startTime, endTime string) (WithdrawalsList, *FwdData, error) {
	result := WithdrawalsList{}
	requestURL := fmt.Sprintf("%s/wapi/v3/withdrawHistory.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return WithdrawalsList{}, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("startTime", startTime).
		WithParam("endTime", endTime).
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
func (bc *Client) Withdraw(symbol, amount, address, name string) (string, *FwdData, error) {
	var result WithdrawResult
	requestURL := fmt.Sprintf("%s/wapi/v3/withdraw.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return "", nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("asset", symbol).
		WithParam("address", address).
		WithParam("name", name).
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

// TransferToMainAccount withdraw from sub account to main account
func (bc *Client) TransferToMainAccount(asset, amount string) (int64, *FwdData, error) {
	var (
		result TransferToMasterResponse
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/sub-account/transfer/subToMaster", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return 0, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("asset", asset).
		WithParam("amount", amount).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return 0, fwd, err
	}
	return result.TxID, fwd, err
}

// SubAccountList list sub account detail
func (bc *Client) SubAccountList(email, status string) (SubAccountResult, *FwdData, error) {
	var (
		result SubAccountResult
	)
	requestURL := fmt.Sprintf("%s/wapi/v3/sub-account/list.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("email", email).
		WithParam("status", status).
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

// TransferToSubHistory list transfer to sub account history
func (bc *Client) TransferToSubHistory(email string, fromTime, toTime int64) (TransferToSubHistoryResult, *FwdData, error) {
	var (
		result TransferToSubHistoryResult
	)
	requestURL := fmt.Sprintf("%s/wapi/v3/sub-account/transfer/history.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("email", email)
	if fromTime != 0 && toTime != 0 {
		rr = rr.WithParam("startTime", strconv.FormatInt(fromTime, 10)).
			WithParam("endTime", strconv.FormatInt(toTime, 10))
	}
	rb := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rb, &result)
	if err != nil {
		return result, fwd, err
	}
	if !result.Success && fwd != nil {
		return result, fwd, errors.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result, fwd, err
}

// AssetTransfer transfer between main <-> sub and sub<->sub
func (bc *Client) AssetTransfer(fromEmail, toEmail, asset string, amount float64) (TransferResult, *FwdData, error) {
	var (
		result TransferResult
	)
	requestURL := fmt.Sprintf("%s/wapi/v3/sub-account/transfer.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("fromEmail", fromEmail).
		WithParam("toEmail", toEmail).
		WithParam("asset", asset).
		WithParam("amount", strconv.FormatFloat(amount, 'f', -1, 64)).
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

// SubAccountAssetBalances transfer between main and sub acc
func (bc *Client) SubAccountAssetBalances(email string) (SubAccountAssetBalancesResult, *FwdData, error) {
	var (
		result SubAccountAssetBalancesResult
	)
	requestURL := fmt.Sprintf("%s/wapi/v3/sub-account/assets.html", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("email", email).
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
