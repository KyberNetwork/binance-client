package binance

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/shopspring/decimal"
)

type StakingProductPositionResponse []StakingProductPosition

type StakingProductPosition struct {
	PositionID          int64           `json:"positionId"`
	ProductID           string          `json:"productId"`
	Asset               string          `json:"asset"`
	Amount              decimal.Decimal `json:"amount"`
	PurchaseTime        int64           `json:"purchaseTime"`
	Duration            int64           `json:"duration"`
	AccrualDays         int64           `json:"accrualDays"`
	RewardAsset         string          `json:"rewardAsset"`
	APY                 decimal.Decimal `json:"apy"`
	RewardAmt           decimal.Decimal `json:"rewardAmt"`
	ExtraRewardAsset    string          `json:"extraRewardAsset"`
	ExtraRewardAPY      decimal.Decimal `json:"extraRewardAPY"`
	EstExtraRewardAmt   decimal.Decimal `json:"estExtraRewardAmt"`
	NextInterestPay     decimal.Decimal `json:"nextInterestPay"`
	NextInterestPayDate int64           `json:"nextInterestPayDate"`
	PayInterestPeriod   int64           `json:"payInterestPeriod"`
	RedeemAmountEarly   decimal.Decimal `json:"redeemAmountEarly"`
	InterestEndDate     int64           `json:"interestEndDate"`
	DeliverDate         int64           `json:"deliverDate"`
	RedeemPeriod        int64           `json:"redeemPeriod"`
	RedeemingAmt        decimal.Decimal `json:"redeemingAmt"`
	CanRedeemEarly      bool            `json:"canRedeemEarly"`
	Renewable           bool            `json:"renewable"`
	Type                string          `json:"type"`
	Status              string          `json:"status"`
}

type StakingProduct string

const (
	StakingProductLockStaking         = "STAKING"
	StakingProductFlexibleDefiStaking = "F_DEFI"
	StakingProductLockDefiStaking     = "L_DEFI"
)

func (bc *Client) GetStakingProductPosition(product StakingProduct, asset string, page, size int) (StakingProductPositionResponse, *FwdData, error) {
	var (
		result StakingProductPositionResponse
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/staking/position", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, nil, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey).
		WithParam("product", string(product))
	if asset != "" {
		rr = rr.WithParam("asset", asset)
	}
	if page > 0 {
		rr = rr.WithParam("current", strconv.Itoa(page))
	}
	if size > 0 {
		rr.WithParam("size", strconv.Itoa(size))
	}
	rq := rr.SignedRequest(bc.secretKey)
	fwd, err := bc.doRequest(rq, &result)
	if err != nil {
		return result, fwd, err
	}
	return result, fwd, err
}
