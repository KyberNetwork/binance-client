package binance

import (
	"fmt"
	"net/http"
	"strconv"
)

// ListenKey is listen for user data stream
type ListenKey struct {
	ListenKey string `json:"listenKey"`
}

const (
	listenKeySpotAPI = "api/v3/userDataStream"
)

// CreateListenKeySpot create a listen key for user data stream
func (bc *Client) CreateListenKeySpot() (string, error) {
	return bc.createListenKey(listenKeySpotAPI)
}

// KeepListenKeyAliveSpot keep it alive
func (bc *Client) KeepListenKeyAliveSpot(listenKey string) error {
	return bc.keepListenKeyAlive(listenKey, listenKeySpotAPI)
}

// GetAccountState return account info
func (bc *Client) GetAccountState() (AccountState, error) {
	var (
		response AccountState
	)
	requestURL := fmt.Sprintf("%s/api/v3/account", bc.apiBaseURL)
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
	requestURL := fmt.Sprintf("%s/api/v3/order", bc.apiBaseURL)
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
	requestURL := fmt.Sprintf("%s/api/v3/openOrders", bc.apiBaseURL)
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
	requestURL := fmt.Sprintf("%s/api/v3/order", bc.apiBaseURL)
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
	requestURL := fmt.Sprintf("%s/api/v3/trades", bc.apiBaseURL)
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
func (bc *Client) GetAccountTradeHistory(symbol, startTime, endTime string, limit int64, fromID string) (AccountTradeHistoryList, *FwdData, error) {
	result := AccountTradeHistoryList{}
	requestURL := fmt.Sprintf("%s/api/v3/myTrades", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("limit", strconv.FormatInt(limit, 10))
	if startTime != "" {
		rr = rr.WithParam("startTime", startTime)
	}
	if endTime != "" {
		rr = rr.WithParam("endTime", endTime)
	}
	if limit != 0 {
		rr = rr.WithParam("limit", strconv.FormatInt(limit, 10))
	}
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
func (bc *Client) WithdrawHistory(coin, startTime, endTime, status string) (WithdrawalsList, *FwdData, error) {
	result := WithdrawalsList{}
	requestURL := fmt.Sprintf("%s/sapi/v1/capital/withdraw/history", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return WithdrawalsList{}, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey)
	if coin != "" {
		rr = rr.WithParam("coin", coin)
	}
	if status != "" {
		rr = rr.WithParam("status", status)
	}
	if startTime != "" {
		rr = rr.WithParam("startTime", startTime)
	}
	if endTime != "" {
		rr = rr.WithParam("endTime", endTime)
	}
	rq := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rq, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

// DepositHistory query recent withdraw list
func (bc *Client) DepositHistory(coin, status, startTime, endTime string) (DepositsList, *FwdData, error) {
	result := DepositsList{}
	requestURL := fmt.Sprintf("%s/sapi/v1/capital/deposit/hisrec", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return DepositsList{}, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey)
	if coin != "" {
		rr = rr.WithParam("coin", coin)
	}
	if startTime != "" {
		rr = rr.WithParam("startTime", startTime)
	}
	if endTime != "" {
		rr = rr.WithParam("endTime", endTime)
	}
	if status != "" {
		rr = rr.WithParam("status", status)
	}
	rq := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rq, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

// CancelOrder cancel an order
func (bc *Client) CancelOrder(symbol string, id int64) (CancelResult, *FwdData, error) {
	result := CancelResult{}
	requestURL := fmt.Sprintf("%s/api/v3/order", bc.apiBaseURL)
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
	requestURL := fmt.Sprintf("%s/api/v3/openOrders", bc.apiBaseURL)
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
func (bc *Client) Withdraw(coin, amount, address, network, name, orderID string) (string, *FwdData, error) {
	var result WithdrawResult
	requestURL := fmt.Sprintf("%s/sapi/v1/capital/withdraw/apply", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return "", nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("coin", coin).
		WithParam("withdrawOrderId", orderID).
		WithParam("network", network).
		WithParam("address", address).
		WithParam("name", name).
		WithParam("amount", amount).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return "", fwd, err
	}
	return result.ID, fwd, err
}

// TransferToMainAccount withdraw from sub account to main account
func (bc *Client) TransferToMainAccount(asset, amount string) (int64, *FwdData, error) {
	var (
		result TransferToMasterResponse
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/sub-account/transfer/subToMaster", bc.apiBaseURL)
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
func (bc *Client) SubAccountList(email, isFreeze string) (SubAccountResult, *FwdData, error) {
	var (
		result SubAccountResult
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/sub-account/list", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey)
	if email != "" {
		rr = rr.WithParam("email", email)
	}
	if isFreeze != "" {
		rr = rr.WithParam("isFreeze", isFreeze)
	}
	rq := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rq, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

// SubAccountTransferHistory list transfer to sub account history
func (bc *Client) SubAccountTransferHistory(fromEmail, toEmail, startTime, endTime string) (SubAccountTransferHistoryResult, *FwdData, error) {
	var (
		result SubAccountTransferHistoryResult
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/sub-account/sub/transfer/history", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("startTime", startTime).
		WithParam("endTime", endTime)
	if fromEmail != "" {
		rr = rr.WithParam("fromEmail", fromEmail)
	}
	if toEmail != "" {
		rr = rr.WithParam("toEmail", toEmail)
	}
	rb := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rb, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

// AssetTransfer transfer between main <-> sub and sub<->sub
func (bc *Client) AssetTransfer(fromEmail, fromAccType, toEmail, toAccountType, asset, amount string) (TransferResult, *FwdData, error) {
	var (
		result TransferResult
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/sub-account/universalTransfer", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("fromEmail", fromEmail).
		WithParam("toEmail", toEmail).
		WithParam("asset", asset).
		WithParam("amount", amount).
		WithParam("fromAccountType", fromAccType).
		WithParam("toAccountType", toAccountType).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	if !result.Success && fwd != nil {
		return result, fwd, fmt.Errorf("binance failure: %s", string(fwd.Data))
	}
	return result, fwd, err
}

// SubAccountAssetBalances transfer between main and sub acc
func (bc *Client) SubAccountAssetBalances(email string) (SubAccountAssetBalancesResult, *FwdData, error) {
	var (
		result SubAccountAssetBalancesResult
	)
	requestURL := fmt.Sprintf("%s/sapi/v3/sub-account/assets", bc.apiBaseURL)
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
	return result, fwd, err
}

// GetDepositAddress ...
func (bc *Client) GetDepositAddress(asset, network string) (BDepositAddress, *FwdData, error) {
	var result BDepositAddress
	requestURL := fmt.Sprintf("%s/sapi/v1/capital/deposit/address", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rq := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("coin", asset)
	if network != "" {
		rq = rq.WithParam("network", network)
	}
	rr := rq.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

// GetAllAssetDetail ...
func (bc *Client) GetAllAssetDetail() (AssetDetailResult, *FwdData, error) {
	var result AssetDetailResult
	requestURL := fmt.Sprintf("%s/sapi/v1/asset/assetDetail", bc.apiBaseURL)
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
	return result, fwd, err
}

// GetExchangeInfo ...
func (bc *Client) GetExchangeInfo() (ExchangeInfo, *FwdData, error) {
	var result ExchangeInfo
	requestURL := fmt.Sprintf("%s/api/v3/exchangeInfo", bc.apiBaseURL)
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
	requestURL := fmt.Sprintf("%s/api/v3/time", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return 0, nil, err
	}
	rr := req.Request()
	fwd, err := bc.doRequest(rr, &result)

	return result.ServerTime, fwd, err
}

// AllCoinInfo return all coin info
func (bc *Client) AllCoinInfo() (AllCoinInfo, *FwdData, error) {
	var result []CoinInfo

	requestURL := fmt.Sprintf("%s/sapi/v1/capital/config/getall", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return nil, fwd, err
	}
	return result, fwd, err
}

// GetOrderBook return order book of a symbol
func (bc *Client) GetOrderBook(symbol, limit string) (OrderBook, *FwdData, error) {
	var result OrderBook

	requestURL := fmt.Sprintf("%s/api/v3/depth", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return OrderBook{}, nil, err
	}
	rr := req.WithParam("symbol", symbol).WithParam("limit", limit).Request()
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return OrderBook{}, fwd, err
	}
	return result, fwd, err
}

// TickerData return ticker data
func (bc *Client) TickerData() ([]TickerEntry, *FwdData, error) {
	var result []TickerEntry
	requestURL := fmt.Sprintf("%s/api/v3/ticker/bookTicker", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.Request()
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

func (bc *Client) GetSubAccountFutureSummary(futuresType, page, limit int) (SubAccountFutureSummaryResponse, *FwdData, error) {
	var result SubAccountFutureSummaryResponse
	requestURL := fmt.Sprintf("%s/sapi/v2/sub-account/futures/accountSummary", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	req = req.WithParam("futuresType", strconv.FormatInt(int64(futuresType), 10))

	if page < 1 {
		page = 1
	}
	req = req.WithParam("page", strconv.FormatInt(int64(page), 10))

	if limit == 0 {
		limit = 10
	}
	req = req.WithParam("limit", strconv.FormatInt(int64(limit), 10))

	rr := req.WithHeader(apiKeyHeader, bc.apiKey).SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}
