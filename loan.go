package binance

import (
	"fmt"
	"github.com/shopspring/decimal"
	"net/http"
)

type LoanRow struct {
	LoanCoin         string          `json:"loanCoin"`
	TotalDebt        decimal.Decimal `json:"totalDebt"`
	CollateralCoin   string          `json:"collateralCoin"`
	CollateralAmount decimal.Decimal `json:"collateralAmount"`
	CurrentLTV       decimal.Decimal `json:"currentLTV"`
}

type LoanData struct {
	Rows  []LoanRow `json:"rows"`
	Total int64     `json:"total"`
}

func (bc *Client) GetFlexibleLoanOnGoingOrder(loanCoin, collateralCoin string, page, limit string) (LoanData, error) {
	var (
		result LoanData
	)
	requestURL := fmt.Sprintf("%s/sapi/v2/loan/flexible/ongoing/orders", bc.apiBaseURL)
	req, err := NewRequestBuilder(http.MethodGet, requestURL, nil)
	if err != nil {
		return result, err
	}
	rr := req.WithHeader(apiKeyHeader, bc.apiKey)
	if loanCoin != "" {
		rr = rr.WithParam("loanCoin", loanCoin)
	}
	if collateralCoin != "" {
		rr = rr.WithParam("collateralCoin", collateralCoin)
	}
	if page != "" {
		rr = rr.WithParam("current", page)
	}
	if limit != "" {
		rr = rr.WithParam("limit", limit)
	} else {
		rr = rr.WithParam("limit", "100")
	}
	s := rr.SignedRequest(bc.secretKey)
	_, err = bc.doRequest(s, &result)
	if err != nil {
		return result, err
	}
	return result, err
}
