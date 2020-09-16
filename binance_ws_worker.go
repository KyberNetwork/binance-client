package binance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/buger/jsonparser"
	ws "github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/KyberNetwork/cex_account_data/common"
	"github.com/KyberNetwork/cex_account_data/lib/caller"
	"github.com/KyberNetwork/cex_account_data/lib/ocache"
)

const (
	balanceUpdate   = "balanceUpdate"
	executionReport = "executionReport"
)

// AccountDataWorker object
type AccountDataWorker struct {
	restClient       *Client
	sugar            *zap.SugaredLogger
	accountInfoStore *common.BinanceAccountInfoStore
	completedOrder   *ocache.OCache
	accountID        string
}

// NewAccountDataWorker create new account worker instance
func NewAccountDataWorker(sugar *zap.SugaredLogger, store *common.BinanceAccountInfoStore, respClient *Client,
	cache *ocache.OCache, id string) *AccountDataWorker {
	return &AccountDataWorker{
		restClient:       respClient,
		sugar:            sugar,
		accountInfoStore: store,
		completedOrder:   cache,
		accountID:        id,
	}
}

func (bc *AccountDataWorker) processMessages(messages chan []byte) {
	var (
		logger = bc.sugar.With("func", caller.GetCurrentFunctionName())
	)
	for m := range messages {
		eventType, err := jsonparser.GetString(m, "e")
		if err != nil {
			logger.Errorw("failed to get eventType", "error", err)
			return
		}
		switch eventType {
		case balanceUpdate:
			var balanceUpdate common.BalanceUpdate
			if err := json.Unmarshal(m, &balanceUpdate); err != nil {
				logger.Errorw("failed to unmarshal balanceUpdate", "error", err)
				return
			}
			balanceUpdateBytes, _ := json.Marshal(balanceUpdate)
			logger.Infow("balance update", "content", fmt.Sprintf("%s", balanceUpdateBytes))
			if err := bc.accountInfoStore.UpdateBalanceDelta(&balanceUpdate); err != nil {
				logger.Errorw("failed to update account balance delta", "error", err)
				return
			}
		case executionReport:
			o, err := parseAccountOrder(m)
			if err != nil {
				logger.Errorw("failed to parse order info", "err", err)
				return
			}
			logger.Infow("update order state", "order_id", o.OrderID,
				"state", o.CurrentOrderStatus, "symbol", o.Symbol)
			oBytes, _ := json.Marshal(o)
			logger.Infow("execution report", "content", fmt.Sprintf("%s", oBytes))
			order, del, err := bc.accountInfoStore.UpdateOrder(o)
			if err != nil {
				logger.Errorw("failed to update order info", "err", err)
				return
			}
			if del {
				bc.completedOrder.Set(common.MakeCompletedOrderID(bc.accountID, order.Symbol, order.OrderID), order)
			}
			if err = bc.updateAccountStateFromRest(); err != nil {
				logger.Errorw("failed to update account from rest")
			}
		}
	}
}

func parseAccountOrder(m []byte) (*common.ExecutionReport, error) {
	e := common.ExecutionReport{}
	var err error
	e.EventTime, err = jsonparser.GetInt(m, "E")
	if err != nil {
		return nil, err
	}
	e.Symbol, err = jsonparser.GetString(m, "s")
	if err != nil {
		return nil, err
	}
	e.ClientOrderID, err = jsonparser.GetString(m, "c")
	if err != nil {
		return nil, err
	}
	e.Side, err = jsonparser.GetString(m, "S")
	if err != nil {
		return nil, err
	}
	e.OrderType, err = jsonparser.GetString(m, "o")
	if err != nil {
		return nil, err
	}
	e.TimeInForce, err = jsonparser.GetString(m, "f")
	if err != nil {
		return nil, err
	}
	e.Quantity, err = jsonparser.GetString(m, "q")
	if err != nil {
		return nil, err
	}
	e.Price, err = jsonparser.GetString(m, "p")
	if err != nil {
		return nil, err
	}
	e.StopPrice, err = jsonparser.GetString(m, "P")
	if err != nil {
		return nil, err
	}
	e.IcebergQuantity, err = jsonparser.GetString(m, "F")
	if err != nil {
		return nil, err
	}
	e.OrderListID, err = jsonparser.GetInt(m, "g")
	if err != nil {
		return nil, err
	}
	e.OriginalClientOrderID, err = jsonparser.GetString(m, "C")
	if err != nil {
		return nil, err
	}
	e.CurrentExecutionType, err = jsonparser.GetString(m, "x")
	if err != nil {
		return nil, err
	}
	e.CurrentOrderStatus, err = jsonparser.GetString(m, "X")
	if err != nil {
		return nil, err
	}
	e.RejectReason, err = jsonparser.GetString(m, "r")
	if err != nil {
		return nil, err
	}
	e.OrderID, err = jsonparser.GetInt(m, "i")
	if err != nil {
		return nil, err
	}
	e.LastExecutedQuantity, err = jsonparser.GetString(m, "l")
	if err != nil {
		return nil, err
	}
	e.CumulativeFilledQuantity, err = jsonparser.GetString(m, "z")
	if err != nil {
		return nil, err
	}
	e.LastExecutedPrice, err = jsonparser.GetString(m, "L")
	if err != nil {
		return nil, err
	}
	e.CommissionAmount, err = jsonparser.GetString(m, "n")
	if err != nil {
		return nil, err
	}
	e.TransactionTime, err = jsonparser.GetInt(m, "T")
	if err != nil {
		return nil, err
	}
	e.TradeID, err = jsonparser.GetInt(m, "t")
	if err != nil {
		return nil, err
	}
	e.OrderCreationTime, err = jsonparser.GetInt(m, "O")
	if err != nil {
		return nil, err
	}
	e.QuoteOrderQty, err = jsonparser.GetString(m, "Q")
	if err != nil {
		return nil, err
	}
	e.CumulativeQuoteAssetTransactedQuantity, err = jsonparser.GetString(m, "Z")
	if err != nil {
		return nil, err
	}
	e.IsOrderInTheBook, err = jsonparser.GetBoolean(m, "w")
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// subscribeDataStream subscribe to a data stream
func (bc *AccountDataWorker) subscribeDataStream(messages chan<- []byte, listenKey string) error {
	var (
		logger   = bc.sugar.With("func", caller.GetCurrentFunctionName())
		wsDialer ws.Dialer
	)
	endpoint := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s", listenKey)
	wsConn, _, err := wsDialer.Dial(endpoint, nil)
	if err != nil {
		logger.Errorw("failed to connect to websocket", "error", err)
		return err
	}
	defer func() {
		_ = wsConn.Close()
	}()
	go func() {
		tick := time.NewTicker(time.Second * 30)
		defer tick.Stop()
		for range tick.C {
			logger.Infow("sending pong..")
			err := wsConn.WriteControl(ws.PongMessage, []byte("pong"), time.Now().Add(time.Second*2))
			if err != nil {
				logger.Errorw("websocket connection error detected", "err", err)
				break
			}
		}
	}()
	tm := time.NewTimer(time.Second)
	tm.Stop()
	for {
		_, m, err := wsConn.ReadMessage()
		if err != nil {
			logger.Errorw("read message error", "err", err)
			return err
		}
		tm.Reset(time.Second)
		select {
		case messages <- m:
			if !tm.Stop() {
				<-tm.C
			}
		case <-tm.C:
			logger.Error("failed to insert message")
		}
	}
}

func (bc *AccountDataWorker) initWSSession() (string, error) {

	listenKey, err := bc.restClient.CreateListenKey()
	if err != nil {
		bc.sugar.Errorw("failed to create listen key", "error", err)
		return "", err
	}
	bc.sugar.Info("fetched listenKey ...", listenKey[len(listenKey)-5:])
	// init account info
	accountState, err := bc.restClient.GetAccountState()
	if err != nil {
		bc.sugar.Errorw("failed to init account info", "error", err)
		return "", err
	}
	orders, err := bc.restClient.GetOpenOrders()
	if err != nil {
		bc.sugar.Errorw("failed to read open orders", "err", err)
		return "", err
	}
	info := &common.BinanceAccountInfo{
		State:     &accountState,
		OpenOrder: make(map[string]*common.OpenOrder),
	}
	for _, o := range orders {
		info.OpenOrder[common.UniqOrder(o.Symbol, o.OrderID)] = o
	}
	bc.accountInfoStore.SetData(info)
	return listenKey, nil
}

// Run the websocket
func (bc *AccountDataWorker) Run() {
	messages := make(chan []byte, 256)
	go bc.processMessages(messages)
	go func() {
		for {
			key, err := bc.initWSSession()
			if err != nil {
				bc.sugar.Errorw("failed to init session", "err", err)
				time.Sleep(time.Second * 3)
				// TODO: consider to clear account data when we cant connect to binance
				continue
			}
			updater := bc.keepAliveKey(key)
			err = bc.subscribeDataStream(messages, key)
			// we got error mostly cause by connection reset, or binance kick us
			if err != nil {
				bc.sugar.Errorw("subscribe data stream broken, retry after seconds", "error", err)
				updater.Stop()
				time.Sleep(time.Second * 5)
			}
		}
	}()
}

func (bc *AccountDataWorker) keepAliveKey(key string) *time.Ticker {
	t := time.NewTicker(time.Minute * 30)
	go func() {
		for range t.C {
			err := bc.restClient.KeepListenKeyAlive(key)
			if err != nil {
				bc.sugar.Errorw("failed to keep listen key alive", "err", err)
			}
		}
	}()
	return t
}

func (bc *AccountDataWorker) updateAccountStateFromRest() error {
	accountState, err := bc.restClient.GetAccountState()
	if err != nil {
		bc.sugar.Errorw("failed to get account state from binance", "error", err)
		return err
	}
	return bc.accountInfoStore.SetAccountState(&accountState)
}
