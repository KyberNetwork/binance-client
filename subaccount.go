package binance

import (
	"fmt"
	"net/http"

	"github.com/shopspring/decimal"
)

type SubAccountMarginAccountDetailResponse struct {
	Email               string          `json:"email"`
	MarginLevel         decimal.Decimal `json:"marginLevel"`
	TotalAssetOfBtc     decimal.Decimal `json:"totalAssetOfBtc"`
	TotalLiabilityOfBtc decimal.Decimal `json:"totalLiabilityOfBtc"`
	TotalNetAssetOfBtc  decimal.Decimal `json:"totalNetAssetOfBtc"`
	MarginTradeCoeffVo  struct {
		ForceLiquidationBar decimal.Decimal `json:"forceLiquidationBar"`
		MarginCallBar       decimal.Decimal `json:"marginCallBar"`
		NormalBar           decimal.Decimal `json:"normalBar"`
	} `json:"marginTradeCoeffVo"`
	MarginUserAssetVoList []struct {
		Asset    string          `json:"asset"`
		Borrowed decimal.Decimal `json:"borrowed"`
		Free     decimal.Decimal `json:"free"`
		Interest decimal.Decimal `json:"interest"`
		Locked   decimal.Decimal `json:"locked"`
		NetAsset decimal.Decimal `json:"netAsset"`
	} `json:"marginUserAssetVoList"`
}

func (bc *Client) GetSubAccountMarginAccountDetail(email string) (SubAccountMarginAccountDetailResponse, *FwdData, error) {
	var (
		result SubAccountMarginAccountDetailResponse
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/sub-account/margin/account", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("email", email).SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rr, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}

type FuturePositionRiskVos struct {
	EntryPrice       decimal.Decimal `json:"entryPrice"`
	Leverage         decimal.Decimal `json:"leverage"`
	MaxNotional      string          `json:"maxNotional"`
	LiquidationPrice decimal.Decimal `json:"liquidationPrice"`
	MarkPrice        decimal.Decimal `json:"markPrice"`
	PositionAmount   decimal.Decimal `json:"positionAmount"`
	Symbol           string          `json:"symbol"`
	UnrealizedProfit decimal.Decimal `json:"unrealizedProfit"`
}

type DeliveryPositionRiskVos struct {
	EntryPrice       decimal.Decimal `json:"entryPrice"`
	MarkPrice        decimal.Decimal `json:"markPrice"`
	Leverage         decimal.Decimal `json:"leverage"`
	Isolated         string          `json:"isolated"`
	IsolatedWallet   string          `json:"isolatedWallet"`
	IsolatedMargin   string          `json:"isolatedMargin"`
	IsAutoAddMargin  string          `json:"isAutoAddMargin"`
	PositionSide     string          `json:"positionSide"`
	PositionAmount   decimal.Decimal `json:"positionAmount"`
	Symbol           string          `json:"symbol"`
	UnrealizedProfit decimal.Decimal `json:"unrealizedProfit"`
}

type SubAccountFuturesPositionRiskResponse struct {
	FuturePositionRiskVos   []FuturePositionRiskVos   `json:"futurePositionRiskVos"`
	DeliveryPositionRiskVos []DeliveryPositionRiskVos `json:"deliveryPositionRiskVos"`
}

func (bc *Client) SubAccountFuturesPositionRisk(email string, futuresType string) (SubAccountFuturesPositionRiskResponse, error) {
	var (
		result SubAccountFuturesPositionRiskResponse
	)
	requestURL := fmt.Sprintf("%s/sapi/v2/sub-account/futures/positionRisk", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("email", email).
		WithParam("futuresType", futuresType).
		SignedRequest(bc.secretKey)
	_, err = bc.doRequest(rr, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

type SubAccountAsset struct {
	Freeze      decimal.Decimal `json:"freeze"`
	Withdrawing decimal.Decimal `json:"withdrawing"`
	Asset       string          `json:"asset"`
	Free        decimal.Decimal `json:"free"`
	Locked      decimal.Decimal `json:"locked"`
}

type SubAccountAssets struct {
	Balances []SubAccountAsset `json:"balances"`
}

func (bc *Client) SubAccountAssets(email string) (SubAccountAssets, error) {
	var (
		result SubAccountAssets
	)
	requestURL := fmt.Sprintf("%s/sapi/v4/sub-account/assets", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("email", email).
		SignedRequest(bc.secretKey)
	_, err = bc.doRequest(rr, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}
