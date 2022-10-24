package binance

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	apiKeyHeader = "X-MBX-APIKEY"

	WalletSpot           = "SPOT"
	WalletIsolatedMargin = "ISOLATED_MARGIN"
)

//go:generate stringer -type=WalletType -linecomment
type WalletType int

const (
	SpotWallet           WalletType = iota + 1 // SPOT
	IsolatedMarginWallet                       // ISOLATED_MARGIN
)

// FwdData contain data we forward to client
type FwdData struct {
	Status      int
	ContentType string
	Data        []byte
}

type APIError struct {
	Code int
	Msg  string
}

func (e APIError) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

func newAPIError(code int, msg string) error {
	return &APIError{
		Code: code,
		Msg:  msg,
	}
}

func ToAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// AccountState is balance state of tokens
type AccountState struct {
	StatusImpl
	MakerCommission  int64     `json:"makerCommission"`
	TakerCommission  int64     `json:"takerCommission"`
	BuyerCommission  int64     `json:"buyerCommission"`
	SellerCommission int64     `json:"sellerCommission"`
	CanTrade         bool      `json:"canTrade"`
	CanWithdraw      bool      `json:"canWithdraw"`
	CanDeposit       bool      `json:"canDeposit"`
	UpdateTime       uint64    `json:"updateTime"`
	AccountType      string    `json:"accountType"`
	Balances         []Balance `json:"balances"`
	Permissions      []string  `json:"permissions"`
}

func (a AccountState) TokensBalance() map[string]Balance {
	res := make(map[string]Balance, len(a.Balances))
	for _, b := range a.Balances {
		res[b.Asset] = b
	}
	return res
}

// BAccountInfo object
type BAccountInfo struct {
	State     *AccountState
	OpenOrder map[string]*OpenOrder
}

// Balance of account
type Balance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

// PayloadBalance is balance object from socket payload
type PayloadBalance struct {
	Asset string `json:"a"`
	Free  string `json:"f"`
	Lock  string `json:"l"`
}

// OutboundAccountPosition object
type OutboundAccountPosition struct {
	EventTime  uint64           `json:"E"`
	LastUpdate uint64           `json:"u"`
	Balance    []PayloadBalance `json:"B"`
}

// BalanceUpdate payload
type BalanceUpdate struct {
	EventTime    int64  `json:"E"`
	Asset        string `json:"a"`
	BalanceDelta string `json:"d"`
	ClearTime    int64  `json:"T"`
}

// ExecutionReport object
type ExecutionReport struct {
	EventTime                              int64  `json:"E"`
	Symbol                                 string `json:"s"`
	ClientOrderID                          string `json:"c"`
	Side                                   string `json:"S"`
	OrderType                              string `json:"o"`
	TimeInForce                            string `json:"f"`
	Quantity                               string `json:"q"`
	Price                                  string `json:"p"`
	StopPrice                              string `json:"P"`
	IcebergQuantity                        string `json:"F"`
	OrderListID                            int64  `json:"g"`
	OriginalClientOrderID                  string `json:"C"`
	CurrentExecutionType                   string `json:"x"`
	CurrentOrderStatus                     string `json:"X"`
	RejectReason                           string `json:"r"`
	OrderID                                int64  `json:"i"`
	LastExecutedQuantity                   string `json:"l"`
	CumulativeFilledQuantity               string `json:"z"`
	LastExecutedPrice                      string `json:"L"`
	CommissionAmount                       string `json:"n"`
	TransactionTime                        int64  `json:"T"`
	TradeID                                int64  `json:"t"`
	OrderCreationTime                      int64  `json:"O"`
	QuoteOrderQty                          string `json:"Q"`
	CumulativeQuoteAssetTransactedQuantity string `json:"Z"`
	IsOrderInTheBook                       bool   `json:"w"`
}

// OpenOrder ...
type OpenOrder struct {
	Symbol              string `json:"symbol"`
	OrderID             int64  `json:"orderId"`
	OrderListID         int64  `json:"orderListId"`
	ClientOrderID       string `json:"clientOrderId"`
	Price               string `json:"price"`
	OrigQty             string `json:"origQty"`
	ExecutedQty         string `json:"executedQty"`
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status              string `json:"status"`
	TimeInForce         string `json:"timeInForce"`
	Type                string `json:"type"`
	Side                string `json:"side"`
	StopPrice           string `json:"stopPrice"`
	IcebergQty          string `json:"icebergQty"`
	Time                int64  `json:"time"`
	UpdateTime          int64  `json:"updateTime"`
	IsWorking           bool   `json:"isWorking"`
	OrigQuoteOrderQty   string `json:"origQuoteOrderQty"`
}

// TradeHistoryList object for recent trade on binance
type TradeHistoryList []struct {
	ID           uint64 `json:"id"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	Time         uint64 `json:"time"`
	IsBuyerMaker bool   `json:"isBuyerMaker"`
	IsBestMatch  bool   `json:"isBestMatch"`
}

// TransferToMasterResponse ...
type TransferToMasterResponse struct {
	TxID int64 `json:"txnId"`
}

// SubAccountResult ...
type SubAccountResult struct {
	SubAccounts []struct {
		Email      string `json:"email"`
		IsFreeze   bool   `json:"isFreeze"`
		CreateTime int64  `json:"createTime"`
	} `json:"subAccounts"`
}

// SubAccountTransferHistoryResult ...
type SubAccountTransferHistoryResult []SubAccountTransferEntry

type SubAccountTransferEntry struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Asset  string `json:"asset"`
	Qty    string `json:"qty"`
	Status string `json:"status"`
	TranId int64  `json:"tranId"`
	Time   int64  `json:"time"`
}

// TransferResult ...
type TransferResult struct {
	StatusImpl
	TxnID string `json:"txnId"`
}

// SubAccountAssetBalancesResult ...
type SubAccountAssetBalancesResult struct {
	Balances []struct {
		Asset  string  `json:"asset"`
		Free   float64 `json:"free"`
		Locked float64 `json:"locked"`
	} `json:"balances"`
}

// BStatus ...
type BStatus interface {
	Status() (bool, string)
}

// StatusImpl ...
type StatusImpl struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}

// Status ...
func (b *StatusImpl) Status() (bool, string) {
	return b.Success, b.Msg
}

// AccountTradeHistoryList object for binance account trade history
type AccountTradeHistoryList []struct {
	Symbol          string `json:"symbol"`
	ID              uint64 `json:"id"`
	OrderID         uint64 `json:"orderId"`
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	QuoteQty        string `json:"quoteQty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Time            uint64 `json:"time"`
	IsBuyer         bool   `json:"isBuyer"`
	IsMaker         bool   `json:"isMaker"`
	IsBestMatch     bool   `json:"isBestMatch"`
}

// WithdrawalsList ...
type WithdrawalsList []WithdrawalEntry

// WithdrawalEntry object for withdraw from binance
type WithdrawalEntry struct {
	ID              string `json:"id"`
	Address         string `json:"address"`
	Amount          string `json:"amount"`
	ApplyTime       string `json:"applyTime"`
	Coin            string `json:"coin"`
	WithdrawOrderId string `json:"withdrawOrderId"`
	Network         string `json:"network"`
	TransferType    int    `json:"transferType"`
	Status          int    `json:"status"`
	TxId            string `json:"txId"`
}

// DepositsList ...
type DepositsList []DepositEntry

// DepositEntry ...
type DepositEntry struct {
	Amount       string `json:"amount"`
	Coin         string `json:"coin"`
	Network      string `json:"network"`
	Status       int    `json:"status"`
	Address      string `json:"address"`
	AddressTag   string `json:"addressTag"`
	TxId         string `json:"txId"`
	InsertTime   int64  `json:"insertTime"`
	TransferType int    `json:"transferType"`
	ConfirmTimes string `json:"confirmTimes"`
}

// CancelResult ...
type CancelResult struct {
	Symbol            string `json:"symbol"`
	OrigClientOrderID string `json:"origClientOrderId"`
	OrderID           uint64 `json:"orderId"`
	ClientOrderID     string `json:"clientOrderId"`
}

// BOrder ..
type BOrder struct {
	StatusImpl
	Symbol        string `json:"symbol"`
	OrderID       uint64 `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	Price         string `json:"price"`
	OrigQty       string `json:"origQty"`
	ExecutedQty   string `json:"executedQty"`
	Status        string `json:"status"`
	TimeInForce   string `json:"timeInForce"`
	Type          string `json:"type"`
	Side          string `json:"side"`
	StopPrice     string `json:"stopPrice"`
	IcebergQty    string `json:"icebergQty"`
	Time          uint64 `json:"time"`
}

// WithdrawResult ...
type WithdrawResult struct {
	ID string `json:"id"`
}

// BDepositAddress ...
type BDepositAddress struct {
	Address    string `json:"address"`
	AddressTag string `json:"addressTag"`
	Coin       string `json:"coin"`
	URL        string `json:"url"`
}

// AssetDetailResult ...
type AssetDetailResult struct {
	AssetDetail map[string]AssetDetail `json:"assetDetail"`
}

// AssetDetail ...
type AssetDetail struct {
	MinWithdrawAmount string  `json:"minWithdrawAmount"`
	DepositStatus     bool    `json:"depositStatus"`
	WithdrawFee       float64 `json:"withdrawFee"`
	WithdrawStatus    bool    `json:"withdrawStatus"`
	DepositTip        string  `json:"depositTip"` // reason if deposit status is false
}

// FilterLimit ...
type FilterLimit struct {
	FilterType  string `json:"filterType"`
	MinPrice    string `json:"minPrice"`
	MaxPrice    string `json:"maxPrice"`
	MinQuantity string `json:"minQty"`
	MaxQuantity string `json:"maxQty"`
	StepSize    string `json:"stepSize"`
	TickSize    string `json:"tickSize"`
	MinNotional string `json:"minNotional"`
}

// BSymbol ...
type BSymbol struct {
	Symbol              string        `json:"symbol"`
	BaseAssetPrecision  int           `json:"baseAssetPrecision"`
	QuoteAssetPrecision int           `json:"quoteAssetPrecision"`
	Filters             []FilterLimit `json:"filters"`
}

// ExchangeInfo ...
type ExchangeInfo struct {
	StatusImpl
	Symbols []BSymbol
}

// ServerTime ...
type ServerTime struct {
	StatusImpl
	ServerTime int64 `json:"serverTime"`
}

// CreateOrderResult ...
type CreateOrderResult struct {
	Symbol              string `json:"symbol"`
	OrderID             int64  `json:"orderId"`
	OrderListID         int64  `json:"orderListId"`
	ClientOrderID       string `json:"clientOrderId"`
	TransactTime        uint64 `json:"transactTime"`
	Price               string `json:"price"`
	OrigQty             string `json:"origQty"`
	ExecutedQty         string `json:"executedQty"`
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status              string `json:"status"`
	TimeInForce         string `json:"timeInForce"`
	Type                string `json:"type"`
	Side                string `json:"side"`
}

// FutureOrder ...
type FutureOrder struct {
	ClientOrderID            string `json:"clientOrderId"`
	CummulativeQuantity      string `json:"cumQty"`
	CummulativeQuoteQuantity string `json:"cumQuote"`
	ExecutedQuantity         string `json:"executedQty"`
	OrderID                  uint64 `json:"orderId"`
	AveragePrice             string `json:"avgPrice"`
	OriginQuantity           string `json:"origQty"`
	Price                    string `json:"price"`
	ReduceOnly               bool   `json:"reduceOnly"`
	Side                     string `json:"side"`
	PositionSide             string `json:"positionSide"`
	Status                   string `json:"status"`
	StopPrice                string `json:"stopPrice"`
	ClosePosition            bool   `json:"closePosition"`
	Symbol                   string `json:"symbol"`
	TimeInForce              string `json:"timeInForce"`
	Type                     string `json:"type"`
	OriginType               string `json:"origType"`
	ActivatePrice            string `json:"activatePrice"`
	PriceRate                string `json:"priceRate"`
	UpdateTime               uint64 `json:"updateTime"`
	WorkingType              string `json:"workingType"`
	PriceProtect             bool   `json:"priceProtect"`
}

// MarginAsset ..
type MarginAsset struct {
	AssetFullName  string `json:"assetFullName"`
	AssetName      string `json:"assetName"`
	IsBorrowable   bool   `json:"isBorrowable"`
	IsMortgageable bool   `json:"isMortgageable"`
	UserMinBorrow  string `json:"userMinBorrow"`
	UserMinRepay   string `json:"userMinRepay"`
}

// MarginPair ..
type MarginPair struct {
	ID            uint64 `json:"id"`
	Symbol        string `json:"symbol"`
	Base          string `json:"base"`
	Quote         string `json:"quote"`
	IsMarginTrade bool   `json:"isMarginTrade"`
	IsBuyAllowed  bool   `json:"isBuyAllowed"`
	IsSellAllowed bool   `json:"isSellAllowed"`
}

// CrossMarginAccountDetails ...
type CrossMarginAccountDetails struct {
	BorrowEnabled       bool   `json:"borrowEnabled"`
	MarginLevel         string `json:"marginLevel"`
	TotalAssetOfBtc     string `json:"totalAssetOfBtc"`
	TotalLiabilityOfBtc string `json:"totalLiabilityOfBtc"`
	TotalNetAssetOfBtc  string `json:"totalNetAssetOfBtc"`
	TradeEnabled        bool   `json:"tradeEnabled"`
	TransferEnabled     bool   `json:"transferEnabled"`
	UserAssets          []struct {
		Asset    string `json:"asset"`
		Borrowed string `json:"borrowed"`
		Free     string `json:"free"`
		Interest string `json:"interest"`
		Locked   string `json:"locked"`
		NetAsset string `json:"netAsset"`
	} `json:"userAssets"`
}

// MaxBorrowableResult ...
type MaxBorrowableResult struct {
	Amount      string `json:"amount"`
	BorrowLimit string `json:"borrowLimit"`
}

// CoinInfo ...
type CoinInfo struct {
	Coin             string `json:"coin"`
	DepositAllEnable bool   `json:"depositAllEnable"`
	Free             string `json:"free"`
	Freeze           string `json:"freeze"`
	IPOable          string `json:"ipoable"`
	IsLegalMoney     bool   `json:"isLegalMoney"`
	Locked           string `json:"locked"`
	Name             string `json:"name"`
	NetworkList      []struct {
		AddressRegex       string `json:"addressRegex"`
		Coin               string `json:"coin"`
		DepositDesc        string `json:"depositDesc"`
		DepositEnable      bool   `json:"depositEnable"`
		IsDefault          bool   `json:"isDefault"`
		MinConfirm         int64  `json:"minConfirm"`
		Name               string `json:"name"`
		Network            string `json:"network"`
		ResetAddressStatus bool   `json:"resetAddressStatus"`
		SpecialTips        string `json:"specialTips"`
		UnLockConfirm      int64  `json:"unLockConfirm"`
		WithdrawDesc       string `json:"withdrawDesc"`
		WithdrawEnable     bool   `json:"withdrawEnable"`
		WithdrawFee        string `json:"withdrawFee"`
		WithdrawMin        string `json:"withdrawMin"`
	} `json:"networkList"`
}

// AllCoinInfo ...
type AllCoinInfo []CoinInfo

// ToMap ...
func (a AllCoinInfo) ToMap() map[string]CoinInfo {
	res := make(map[string]CoinInfo, len(a))
	for _, v := range a {
		res[v.Coin] = v
	}
	return res
}

// OrderBook represent orderbook on binance
type OrderBook struct {
	Bids           []RateAndQty `json:"bids"`
	Asks           []RateAndQty `json:"asks"`
	LatestUpdateID int64        `json:"lastUpdateId"`
}

// RateAndQty is price item
type RateAndQty struct {
	Quantity string `json:"quantity"`
	Rate     string `json:"rate"`
}

// UnmarshalJSON custom unmarshal for binaprice
func (bp *RateAndQty) UnmarshalJSON(text []byte) error {
	temp := []interface{}{&bp.Quantity, &bp.Rate}
	if err := json.Unmarshal(text, &temp); err != nil {
		return err
	}
	return nil
}

// TickerEntry ...
type TickerEntry struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	BidQty   string `json:"bidQty"`
	AskPrice string `json:"askPrice"`
	AskQty   string `json:"askQty"`
}

// IsolatedMarginAsset ...
type IsolatedMarginAsset struct {
	Asset         string `json:"asset"`
	BorrowEnabled bool   `json:"borrowEnabled"`
	Borrowed      string `json:"borrowed"`
	Free          string `json:"free"`
	Interest      string `json:"interest"`
	Locked        string `json:"locked"`
	NetAsset      string `json:"netAsset"`
	NetAssetOfBtc string `json:"netAssetOfBtc"`
	RepayEnabled  bool   `json:"repayEnabled"`
	TotalAsset    string `json:"totalAsset"`
}

// IsolatedMarginAssetInfo ...
type IsolatedMarginAssetInfo struct {
	BaseAsset         IsolatedMarginAsset `json:"baseAsset"`
	QuoteAsset        IsolatedMarginAsset `json:"quoteAsset"`
	Symbol            string              `json:"symbol"`
	IsolatedCreated   bool                `json:"isolatedCreated"`
	MarginLevel       string              `json:"marginLevel"`
	MarginLevelStatus string              `json:"marginLevelStatus"`
	MarginRatio       string              `json:"marginRatio"`
	IndexPrice        string              `json:"indexPrice"`
	LiquidatePrice    string              `json:"liquidatePrice"`
	LiquidateRate     string              `json:"liquidateRate"`
	TradeEnabled      bool                `json:"tradeEnabled"`
}

// IsolatedMarginAccountDetails ...
type IsolatedMarginAccountDetails struct {
	Assets              []IsolatedMarginAssetInfo `json:"assets"`
	TotalAssetOfBtc     string                    `json:"totalAssetOfBtc"`
	TotalLiabilityOfBtc string                    `json:"totalLiabilityOfBtc"`
	TotalNetAssetOfBtc  string                    `json:"totalNetAssetOfBtc"`
}

// USDFutureAccountSummary https://binance-docs.github.io/apidocs/spot/en/#get-summary-of-sub-account-39-s-futures-account-v2-for-master-account
type USDFutureAccountSummary struct {
	TotalInitialMargin          decimal.Decimal          `json:"totalInitialMargin"`
	TotalMaintenanceMargin      decimal.Decimal          `json:"totalMaintenanceMargin"`
	TotalMarginBalance          decimal.Decimal          `json:"totalMarginBalance"`
	TotalOpenOrderInitialMargin decimal.Decimal          `json:"totalOpenOrderInitialMargin"`
	TotalPositionInitialMargin  decimal.Decimal          `json:"totalPositionInitialMargin"`
	TotalUnrealizedProfit       decimal.Decimal          `json:"totalUnrealizedProfit"`
	TotalWalletBalance          decimal.Decimal          `json:"totalWalletBalance"`
	Asset                       string                   `json:"asset"`
	SubAccountList              []USDFutureAccountDetail `json:"subAccountList"`
}

type USDFutureAccountDetail struct {
	Email                       string          `json:"email"`
	TotalInitialMargin          decimal.Decimal `json:"totalInitialMargin"`
	TotalMaintenanceMargin      decimal.Decimal `json:"totalMaintenanceMargin"`
	TotalMarginBalance          decimal.Decimal `json:"totalMarginBalance"`
	TotalOpenOrderInitialMargin decimal.Decimal `json:"totalOpenOrderInitialMargin"`
	TotalPositionInitialMargin  decimal.Decimal `json:"totalPositionInitialMargin"`
	TotalUnrealizedProfit       decimal.Decimal `json:"totalUnrealizedProfit"`
	TotalWalletBalance          decimal.Decimal `json:"totalWalletBalance"`
	Asset                       string          `json:"asset"`
}

type CoinFutureAccountDetail struct {
	Email                 string          `json:"email"`
	TotalMarginBalance    decimal.Decimal `json:"totalMarginBalance"`
	TotalUnrealizedProfit decimal.Decimal `json:"totalUnrealizedProfit"`
	TotalWalletBalance    decimal.Decimal `json:"totalWalletBalance"`
	Asset                 string          `json:"asset"`
}

type CoinFutureAccountSummary struct {
	TotalMarginBalanceOfBTC    decimal.Decimal           `json:"totalMarginBalanceOfBTC"`
	TotalUnrealizedProfitOfBTC decimal.Decimal           `json:"totalUnrealizedProfitOfBTC"`
	TotalWalletBalanceOfBTC    decimal.Decimal           `json:"totalWalletBalanceOfBTC"`
	Asset                      string                    `json:"asset"`
	SubAccountList             []CoinFutureAccountDetail `json:"subAccountList"`
}

type SubAccountFutureSummaryResponse struct {
	USDFutureAccountSummary  USDFutureAccountSummary  `json:"futureAccountSummaryResp"`
	CoinFutureAccountSummary CoinFutureAccountSummary `json:"deliveryAccountSummaryResp"`
}

type SubAccountFutureDetailsResponse struct {
	FutureAccountResp struct {
		Email  string `json:"email"`
		Assets []struct {
			Asset                  string          `json:"asset"`
			InitialMargin          decimal.Decimal `json:"initialMargin"`
			MaintenanceMargin      decimal.Decimal `json:"maintenanceMargin"`
			MarginBalance          decimal.Decimal `json:"marginBalance"`
			MaxWithdrawAmount      decimal.Decimal `json:"maxWithdrawAmount"`
			OpenOrderInitialMargin decimal.Decimal `json:"openOrderInitialMargin"`
			PositionInitialMargin  decimal.Decimal `json:"positionInitialMargin"`
			UnrealizedProfit       decimal.Decimal `json:"unrealizedProfit"`
			WalletBalance          decimal.Decimal `json:"walletBalance"`
		} `json:"assets"`
		CanDeposit                  bool            `json:"canDeposit"`
		CanTrade                    bool            `json:"canTrade"`
		CanWithdraw                 bool            `json:"canWithdraw"`
		FeeTier                     int             `json:"feeTier"`
		MaxWithdrawAmount           decimal.Decimal `json:"maxWithdrawAmount"`
		TotalInitialMargin          decimal.Decimal `json:"totalInitialMargin"`
		TotalMaintenanceMargin      decimal.Decimal `json:"totalMaintenanceMargin"`
		TotalMarginBalance          decimal.Decimal `json:"totalMarginBalance"`
		TotalOpenOrderInitialMargin decimal.Decimal `json:"totalOpenOrderInitialMargin"`
		TotalPositionInitialMargin  decimal.Decimal `json:"totalPositionInitialMargin"`
		TotalUnrealizedProfit       decimal.Decimal `json:"totalUnrealizedProfit"`
		TotalWalletBalance          decimal.Decimal `json:"totalWalletBalance"`
		UpdateTime                  int64           `json:"updateTime"`
	} `json:"futureAccountResp"`
	DeliveryAccountResp struct {
		Email  string `json:"email"`
		Assets []struct {
			Asset                  string          `json:"asset"`
			InitialMargin          decimal.Decimal `json:"initialMargin"`
			MaintenanceMargin      decimal.Decimal `json:"maintenanceMargin"`
			MarginBalance          decimal.Decimal `json:"marginBalance"`
			MaxWithdrawAmount      decimal.Decimal `json:"maxWithdrawAmount"`
			OpenOrderInitialMargin decimal.Decimal `json:"openOrderInitialMargin"`
			PositionInitialMargin  decimal.Decimal `json:"positionInitialMargin"`
			UnrealizedProfit       decimal.Decimal `json:"unrealizedProfit"`
			WalletBalance          decimal.Decimal `json:"walletBalance"`
		} `json:"assets"`
		CanDeposit  bool  `json:"canDeposit"`
		CanTrade    bool  `json:"canTrade"`
		CanWithdraw bool  `json:"canWithdraw"`
		FeeTier     int   `json:"feeTier"`
		UpdateTime  int64 `json:"updateTime"`
	} `json:"deliveryAccountResp"`
}
