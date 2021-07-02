package binance

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	listenKeyTypeMarginAPI         = "sapi/v1/userDataStream"
	listenKeyTypeIsolatedMarginAPI = "sapi/v1/userDataStream/isolated"
)

// CreateListenKeyMargin create a listen key for user data stream
func (bc *Client) CreateListenKeyMargin() (string, error) {
	return bc.createListenKey(listenKeyTypeMarginAPI)
}

// KeepListenKeyAliveMargin keep it alive
func (bc *Client) KeepListenKeyAliveMargin(listenKey string) error {
	return bc.keepListenKeyAlive(listenKey, listenKeyTypeMarginAPI)
}

// CreateListenKeyIsolatedMargin create a listen key for user data stream
func (bc *Client) CreateListenKeyIsolatedMargin() (string, error) {
	return bc.createListenKey(listenKeyTypeIsolatedMarginAPI)
}

// KeepListenKeyAliveIsolatedMargin keep it alive
func (bc *Client) KeepListenKeyAliveIsolatedMargin(listenKey string) error {
	return bc.keepListenKeyAlive(listenKey, listenKeyTypeIsolatedMarginAPI)
}

type marginCommonResult struct {
	TranID uint64 `json:"tranId"`
}

// TransferCrossMargin transfer between spot account and cross margin account.
func (bc *Client) TransferCrossMargin(asset, amount string, spotToMargin bool) (uint64, *FwdData, error) {
	transType := "2"
	if spotToMargin {
		transType = "1"
	}
	var (
		result marginCommonResult
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/margin/transfer", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return 0, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("asset", asset).
		WithParam("amount", amount).
		WithParam("type", transType).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return 0, fwd, err
	}

	return result.TranID, fwd, err
}

// Borrow borrow margin
func (bc *Client) Borrow(asset, symbol, amount string, isIsolated bool) (uint64, *FwdData, error) {
	var (
		result marginCommonResult
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/margin/loan", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return 0, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("asset", asset).
		WithParam("amount", amount)
	if isIsolated {
		rr = rr.WithParam("isIsolated", "TRUE").
			WithParam("symbol", symbol)
	}
	sr := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(sr, &result)
	if err != nil {
		return 0, fwd, err
	}
	return result.TranID, fwd, err
}

// Repay repay margin
func (bc *Client) Repay(asset, symbol, amount string, isIsolated bool) (uint64, *FwdData, error) {
	var (
		result marginCommonResult
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/margin/repay", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return 0, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("asset", asset).
		WithParam("amount", amount)
	if isIsolated {
		rr = rr.WithParam("isIsolated", "TRUE").
			WithParam("symbol", symbol)
	}
	sr := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(sr, &result)
	if err != nil {
		return 0, fwd, err
	}
	return result.TranID, fwd, err
}

// GetMarginAsset return asset info
func (bc *Client) GetMarginAsset(asset string) (MarginAsset, *FwdData, error) {
	var (
		result MarginAsset
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/margin/asset", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).WithParam("asset", asset).Request()
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

// GetMarginPair return pair info
func (bc *Client) GetMarginPair(symbol string) (MarginAsset, *FwdData, error) {
	var (
		result MarginAsset
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/margin/pair", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).WithParam("symbol", symbol).Request()
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

// GetAllMarginAssets return all margin assets
func (bc *Client) GetAllMarginAssets() ([]MarginAsset, *FwdData, error) {
	var (
		result []MarginAsset
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/margin/allAssets", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).Request()
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

// GetCrossMarginAccountDetails return margin account details
func (bc *Client) GetCrossMarginAccountDetails() (CrossMarginAccountDetails, *FwdData, error) {
	var (
		result CrossMarginAccountDetails
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/margin/account", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

// GetMaxBorrowable return max borrowable
func (bc *Client) GetMaxBorrowable(asset, isolatedSymbol string) (MaxBorrowableResult, *FwdData, error) {
	var (
		result MaxBorrowableResult
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/margin/maxBorrowable", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("asset", asset).
		WithParam("isolatedSymbol", isolatedSymbol).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

// TransferIsolatedMargin transfer between spot account and isolated margin account.
func (bc *Client) TransferIsolatedMargin(asset, symbol, amount string, transFrom, transferTo WalletType) (uint64, *FwdData, error) {
	var (
		result marginCommonResult
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/margin/isolated/transfer", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return 0, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("asset", asset).
		WithParam("symbol", symbol).
		WithParam("transFrom", transFrom.String()).
		WithParam("transTo", transferTo.String()).
		WithParam("amount", amount).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return 0, fwd, err
	}
	return result.TranID, fwd, err
}

// GetIsolatedMarginAccountDetails return isolated account details
func (bc *Client) GetIsolatedMarginAccountDetails(symbols []string) (IsolatedMarginAccountDetails, *FwdData, error) {
	var (
		result IsolatedMarginAccountDetails
	)
	if len(symbols) > 5 {
		return result, nil, fmt.Errorf("the api only supports max 5 symbols")
	}
	requestURL := fmt.Sprintf("%s/sapi/v1/margin/isolated/account", apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return IsolatedMarginAccountDetails{}, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey)
	if len(symbols) > 0 {
		rr.WithParam("symbols", strings.Join(symbols, ","))
	}
	fwd, err := bc.doRequest(rr.SignedRequest(bc.secretKey), &result)
	if err != nil {
		return IsolatedMarginAccountDetails{}, fwd, err
	}
	return result, fwd, err
}
