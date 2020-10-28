package binance

import (
	"fmt"
	"sync"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/KyberNetwork/cex_account_data/common"
)

// BAccountInfoStore store account info
type BAccountInfoStore struct {
	mu          sync.Mutex
	AccountInfo *BAccountInfo
	l           *zap.SugaredLogger
}

// NewBinanceAccountInfoStore return new binance account info object
func NewBinanceAccountInfoStore(l *zap.SugaredLogger, infoStorage *BAccountInfo) *BAccountInfoStore {
	return &BAccountInfoStore{
		AccountInfo: infoStorage,
		l:           l,
	}
}

// SetAccountState init account info when start the service
func (ai *BAccountInfoStore) SetAccountState(data *AccountState) error {

	ai.mu.Lock()
	defer ai.mu.Unlock()
	ai.AccountInfo.State = data
	return nil
}

// UpdateOrder will update order state
func (ai *BAccountInfoStore) UpdateOrder(o *ExecutionReport) (*OpenOrder, bool, error) {
	orderID := common.MakeCompletedOrderID(o.Symbol, o.OrderID)
	ai.mu.Lock()
	defer ai.mu.Unlock()
	order, ok := ai.AccountInfo.OpenOrder[orderID]
	if !ok {
		order = &OpenOrder{}
		ai.AccountInfo.OpenOrder[orderID] = order
	}

	order.Symbol = o.Symbol
	order.OrderID = o.OrderID
	order.OrderListID = o.OrderListID
	order.ClientOrderID = o.ClientOrderID
	order.Price = o.Price
	order.OrigQty = o.Quantity
	order.ExecutedQty = o.CumulativeFilledQuantity                       // ?
	order.CummulativeQuoteQty = o.CumulativeQuoteAssetTransactedQuantity // ?
	order.Status = o.CurrentOrderStatus
	order.TimeInForce = o.TimeInForce
	order.Type = o.OrderType
	order.Side = o.Side
	order.StopPrice = o.StopPrice
	order.IcebergQty = o.IcebergQuantity
	order.Time = o.OrderCreationTime
	order.UpdateTime = o.EventTime
	order.IsWorking = o.IsOrderInTheBook
	order.OrigQuoteOrderQty = ""

	if o.CurrentOrderStatus == "CANCELED" || o.CurrentOrderStatus == "FILLED" { // "NEW" || o.CurrentOrderStatus == "PARTIALLY_FILLED" || o.CurrentOrderStatus == "PENDING_CANCEL"" {
		delete(ai.AccountInfo.OpenOrder, orderID)
		return order, true, nil
	}
	return order, false, nil
}

// OpenOrders get all open orders
func (ai *BAccountInfoStore) OpenOrders() []*OpenOrder {
	res := make([]*OpenOrder, 0)
	ai.mu.Lock()
	defer ai.mu.Unlock()
	for _, o := range ai.AccountInfo.OpenOrder {
		res = append(res, o)
	}
	return res
}

// OrderStatus get order
func (ai *BAccountInfoStore) OrderStatus(symbol string, orderID int64) (*OpenOrder, error) {

	ai.mu.Lock()
	defer ai.mu.Unlock()
	for _, o := range ai.AccountInfo.OpenOrder {
		if o.Symbol == symbol && o.OrderID == orderID {
			return o, nil
		}
	}
	return nil, fmt.Errorf("order %d for symbol %s not found", orderID, symbol)
}

// UpdateBalanceDelta update balance
func (ai *BAccountInfoStore) UpdateBalanceDelta(balanceUpdate *BalanceUpdate) error {

	ai.mu.Lock()
	defer ai.mu.Unlock()

	for index, balance := range ai.AccountInfo.State.Balances {
		if balance.Asset != balanceUpdate.Asset {
			continue
		}
		delta, err := decimal.NewFromString(balanceUpdate.BalanceDelta)
		if err != nil {
			ai.l.Error("failed to parse balance delta", "err", err)
			return err
		}
		oldBalance, err := decimal.NewFromString(ai.AccountInfo.State.Balances[index].Free)
		if err != nil {
			ai.l.Error("failed to parse old free balance", "err", err)
			return err
		}
		ai.AccountInfo.State.Balances[index].Free = oldBalance.Add(delta).String()
		break
	}
	return nil
}

// UpdateBalance of token
func (ai *BAccountInfoStore) UpdateBalance(balance []*PayloadBalance) error {

	ai.mu.Lock()
	defer ai.mu.Unlock()

	for _, tokenBalance := range balance {
		exist := false
		for index, balance := range ai.AccountInfo.State.Balances {
			if tokenBalance.Asset != balance.Asset {
				continue
			}
			ai.AccountInfo.State.Balances[index].Free = tokenBalance.Free
			ai.AccountInfo.State.Balances[index].Locked = tokenBalance.Lock
			exist = true
			break
		}
		if !exist {
			ai.AccountInfo.State.Balances = append(ai.AccountInfo.State.Balances, Balance{
				Asset:  tokenBalance.Asset,
				Free:   tokenBalance.Free,
				Locked: tokenBalance.Lock,
			})
		}
	}
	return nil
}

// GetAccountState return binance account info
func (ai *BAccountInfoStore) GetAccountState() *AccountState {
	ai.mu.Lock()
	defer ai.mu.Unlock()
	src := ai.AccountInfo.State
	if src == nil {
		return nil
	}
	res := &AccountState{
		MakerCommission:  src.MakerCommission,
		TakerCommission:  src.TakerCommission,
		BuyerCommission:  src.BuyerCommission,
		SellerCommission: src.SellerCommission,
		CanTrade:         src.CanTrade,
		CanWithdraw:      src.CanWithdraw,
		CanDeposit:       src.CanDeposit,
		UpdateTime:       src.UpdateTime,
		AccountType:      src.AccountType,
		Balances:         make([]Balance, len(src.Balances)),
		Permissions:      make([]string, len(src.Permissions)),
	}
	copy(res.Balances, src.Balances)
	copy(res.Permissions, src.Permissions)
	return res
}

// SetData set full data at session init
func (ai *BAccountInfoStore) SetData(info *BAccountInfo) {
	ai.mu.Lock()
	defer ai.mu.Unlock()
	ai.AccountInfo = info
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
	StatusImpl
	SubAccounts []struct {
		Email      string `json:"email"`
		Status     string `json:"status"`
		Activated  bool   `json:"activated"`
		Mobile     string `json:"mobile"`
		GAuth      bool   `json:"gAuth"`
		CreateTime int64  `json:"createTime"`
	} `json:"subAccounts"`
}

// SubAccountTransferHistoryResult ...
type SubAccountTransferHistoryResult struct {
	StatusImpl
	Transfers []struct {
		From  string `json:"from"`
		To    string `json:"to"`
		Asset string `json:"asset"`
		Qty   string `json:"qty"`
		Time  int64  `json:"time"`
	} `json:"transfers"`
}

// TransferResult ...
type TransferResult struct {
	StatusImpl
	TxnID string `json:"txnId"`
}

// SubAccountAssetBalancesResult ...
type SubAccountAssetBalancesResult struct {
	StatusImpl
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
type WithdrawalsList struct {
	StatusImpl
	Withdrawals []WithdrawalEntry `json:"withdrawList"`
}

// WithdrawalEntry object for withdraw from binance
type WithdrawalEntry struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Address   string  `json:"address"`
	Asset     string  `json:"asset"`
	TxID      string  `json:"txId"`
	ApplyTime uint64  `json:"applyTime"`
	Fee       float64 `json:"transactionFee"`
	Status    int     `json:"status"`
}

// DepositsList ...
type DepositsList struct {
	StatusImpl
	Deposits []DepositEntry `json:"depositList"`
}

// DepositEntry ...
type DepositEntry struct {
	InsertTime uint64  `json:"insertTime"`
	Amount     float64 `json:"amount"`
	Asset      string  `json:"asset"`
	Address    string  `json:"address"`
	TxID       string  `json:"txId"`
	Status     int     `json:"status"`
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
	StatusImpl
	ID string `json:"id"`
}

// BDepositAddress ...
type BDepositAddress struct {
	StatusImpl
	Address    string `json:"address"`
	AddressTag string `json:"addressTag"`
	Asset      string `json:"asset"`
}

// AssetDetailResult ...
type AssetDetailResult struct {
	StatusImpl
	AssetDetail AssetDetail `json:"assetDetail"`
}

// AssetDetail ...
type AssetDetail struct {
	MinWithdrawAmount float64 `json:"minWithdrawAmount"`
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
