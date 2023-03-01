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
