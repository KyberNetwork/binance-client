package binance

import (
	"fmt"
	"github.com/shopspring/decimal"
	"net/http"
)

type FlexibleEarnRow struct {
	TotalAmount                    decimal.Decimal            `json:"totalAmount"`
	TierAnnualPercentageRate       map[string]decimal.Decimal `json:"tierAnnualPercentageRate"`
	LatestAnnualPercentageRate     decimal.Decimal            `json:"latestAnnualPercentageRate"`
	YesterdayAirdropPercentageRate decimal.Decimal            `json:"yesterdayAirdropPercentageRate"`
	Asset                          string                     `json:"asset"`
	AirDropAsset                   string                     `json:"airDropAsset"`
	CanRedeem                      bool                       `json:"canRedeem"`
	CollateralAmount               decimal.Decimal            `json:"collateralAmount"`
	ProductId                      string                     `json:"productId"`
	YesterdayRealTimeRewards       decimal.Decimal            `json:"yesterdayRealTimeRewards"`
	CumulativeBonusRewards         decimal.Decimal            `json:"cumulativeBonusRewards"`
	CumulativeRealTimeRewards      decimal.Decimal            `json:"cumulativeRealTimeRewards"`
	CumulativeTotalRewards         decimal.Decimal            `json:"cumulativeTotalRewards"`
	AutoSubscribe                  bool                       `json:"autoSubscribe"`
}

type FlexibleEarnData struct {
	Rows  []FlexibleEarnRow `json:"rows"`
	Total int64             `json:"total"`
}

func (bc *Client) GetSimpleEarnFlexiblePosition(asset, productID string, page, limit string) (FlexibleEarnData, error) {
	var (
		result FlexibleEarnData
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/simple-earn/flexible/position", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey)
	if asset != "" {
		rr = rr.WithParam("asset", asset)
	}
	if productID != "" {
		rr = rr.WithParam("productId", productID)
	}
	if page != "" {
		rr = rr.WithParam("current", page)
	}
	if limit != "" {
		rr = rr.WithParam("size", limit)
	} else {
		rr = rr.WithParam("size", "100")
	}
	s := rr.SignedRequest(bc.secretKey)
	_, err = bc.doRequest(s, &result)
	if err != nil {
		return result, err
	}
	return result, err
}

type LockedEarnRow struct {
	PositionId            int64           `json:"positionId"`
	ParentPositionId      int64           `json:"parentPositionId"`
	ProjectId             string          `json:"projectId"`
	Asset                 string          `json:"asset"`
	Amount                decimal.Decimal `json:"amount"`
	PurchaseTime          decimal.Decimal `json:"purchaseTime"`
	Duration              decimal.Decimal `json:"duration"`
	AccrualDays           decimal.Decimal `json:"accrualDays"`
	RewardAsset           string          `json:"rewardAsset"`
	APY                   string          `json:"APY"`
	RewardAmt             decimal.Decimal `json:"rewardAmt"`
	ExtraRewardAsset      string          `json:"extraRewardAsset"`
	ExtraRewardAPR        decimal.Decimal `json:"extraRewardAPR"`
	EstExtraRewardAmt     string          `json:"estExtraRewardAmt"`
	NextPay               string          `json:"nextPay"`
	NextPayDate           int64           `json:"nextPayDate"`
	PayPeriod             int64           `json:"payPeriod"`
	RedeemAmountEarly     string          `json:"redeemAmountEarly"`
	RewardsEndDate        int64           `json:"rewardsEndDate"`
	DeliverDate           int64           `json:"deliverDate"`
	RedeemPeriod          int64           `json:"redeemPeriod"`
	RedeemingAmt          string          `json:"redeemingAmt"`
	RedeemTo              string          `json:"redeemTo"`
	PartialAmtDeliverDate int64           `json:"partialAmtDeliverDate"`
	CanRedeemEarly        bool            `json:"canRedeemEarly"`
	CanFastRedemption     bool            `json:"canFastRedemption"`
	AutoSubscribe         bool            `json:"autoSubscribe"`
	Type                  string          `json:"type"`
	Status                string          `json:"status"`
	CanReStake            bool            `json:"canReStake"`
}

type LockedEarnData struct {
	Rows  []LockedEarnRow `json:"rows"`
	Total int64           `json:"total"`
}

func (bc *Client) GetSimpleEarnLockedPosition(asset, positionID, projectID string, page, limit string) (LockedEarnData, error) {
	var (
		result LockedEarnData
	)
	requestURL := fmt.Sprintf("%s/sapi/v1/simple-earn/locked/position", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey)
	if asset != "" {
		rr = rr.WithParam("asset", asset)
	}
	if positionID != "" {
		rr = rr.WithParam("positionId", positionID)
	}
	if projectID != "" {
		rr = rr.WithParam("projectId", projectID)
	}
	if page != "" {
		rr = rr.WithParam("current", page)
	}
	if limit != "" {
		rr = rr.WithParam("size", limit)
	} else {
		rr = rr.WithParam("size", "100")
	}
	s := rr.SignedRequest(bc.secretKey)
	_, err = bc.doRequest(s, &result)
	if err != nil {
		return result, err
	}
	return result, err
}
