package binance

import (
	"fmt"
	"net/http"
	"strconv"
)

// CreateFutureOrder ...
func (bc *Client) CreateFutureOrder(symbol, side, positionSide, tradeType, timeInForce, reduceOnly, newClientOrderID, closePosition, workingType, priceProtect, newOrderRespType string,
	price, stopPrice, activationPrice, callbackRate, quantity float64) (FutureOrder, error) {
	var (
		response FutureOrder
	)
	requestURL := fmt.Sprintf("%s/fapi/v1/order", bc.futureAPIBaseURL)
	req, err := NewRequestBuilder(http.MethodPost, requestURL, nil)
	if err != nil {
		return response, err
	}
	quantityStr := strconv.FormatFloat(quantity, 'f', -1, 64)
	priceStr := strconv.FormatFloat(price, 'f', -1, 64)
	rrb := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("symbol", symbol).
		WithParam("side", side).
		WithParam("type", tradeType).
		WithParam("positionSide", positionSide).
		WithParam("quantity", quantityStr).
		WithParam("price", priceStr)
	if timeInForce != "" {
		rrb = rrb.WithParam("timeInForce", timeInForce)
	}
	if newClientOrderID != "" {
		rrb = rrb.WithParam("newClientOrderId", newClientOrderID)
	}
	if closePosition != "" {
		rrb = rrb.WithParam("closePosition", closePosition)
	}
	if workingType != "" {
		rrb = rrb.WithParam("workingType", workingType)
	}
	if priceProtect != "" {
		rrb = rrb.WithParam("priceProtect", priceProtect)
	}
	if newOrderRespType != "" {
		rrb = rrb.WithParam("newOrderRespType", newOrderRespType)
	}
	if stopPrice != 0 {
		stopPriceStr := strconv.FormatFloat(stopPrice, 'f', -1, 64)
		rrb = rrb.WithParam("stopPrice", stopPriceStr)
	}
	if activationPrice != 0 {
		activationPriceStr := strconv.FormatFloat(activationPrice, 'f', -1, 64)
		rrb = rrb.WithParam("activationPrice", activationPriceStr)
	}
	if callbackRate != 0 {
		callbackRateStr := strconv.FormatFloat(callbackRate, 'f', -1, 64)
		rrb = rrb.WithParam("callbackRate", callbackRateStr)
	}
	rr := rrb.SignedRequest(bc.secretKey)
	_, err = bc.doRequest(rr, &response)
	return response, err
}

// PositionInformation ...
type PositionInformation struct {
	EntryPrice       string `json:"entryPrice"`
	MarginType       string `json:"marginType"`
	IsAutoAddMargin  string `json:"isAutoAddMargin"`
	IsolatedMargin   string `json:"isolatedMargin"`
	Leverage         string `json:"leverage"`
	LiquidationPrice string `json:"liquidationPrice"`
	MarkPrice        string `json:"markPrice"`
	MaxNotionalValue string `json:"maxNotionalValue"`
	PositionAmt      string `json:"positionAmt"`
	Symbol           string `json:"symbol"`
	UnrealizedProfit string `json:"unRealizedProfit"`
	PositionSide     string `json:"positionSide"`
}

// GetPositionInformation ...
func (bc *Client) GetPositionInformation(symbol string) ([]PositionInformation, error) {
	var (
		response []PositionInformation
		rr       *http.Request
	)
	requestURL := fmt.Sprintf("%s/fapi/v2/positionRisk", bc.futureAPIBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return response, nil
	}
	if symbol == "" {
		rr = req.WithHeader(apiKeyHeader, bc.apiKey).SignedRequest(bc.secretKey)
	} else {
		rr = req.WithHeader(apiKeyHeader, bc.apiKey).
			WithParam("symbol", symbol).
			SignedRequest(bc.secretKey)
	}
	_, err = bc.doRequest(rr, &response)
	return response, err
}

// FutureAccountBalance ...
type FutureAccountBalance struct {
	AccountAlias       string `json:"accountAlias"`
	Asset              string `json:"asset"`
	Balance            string `json:"balance"`
	CrossWalletBalance string `json:"crossWalletBalance"`
	CrossUnPNL         string `json:"crossUnPnl"`
	AvailableBalance   string `json:"availableBalance"`
	MaxWithdrawAmount  string `json:"maxWithdrawAmount"`
	MarginAvailable    bool   `json:"marginAvailable"`
	UpdateTime         uint64 `json:"updateTime"`
}

// FutureAccountBalance ...
func (bc *Client) FutureAccountBalance() ([]FutureAccountBalance, *FwdData, error) {
	var (
		response []FutureAccountBalance
	)
	requestURL := fmt.Sprintf("%s/fapi/v2/balance", bc.futureAPIBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return response, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &response)
	return response, fwd, err
}
