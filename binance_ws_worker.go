package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/buger/jsonparser"
	ws "github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/KyberNetwork/cex_account_data/common"
	"github.com/KyberNetwork/cex_account_data/lib/caller"
)

const (
	outboundAccountInfo     = "outboundAccountInfo"
	outboundAccountPosition = "outboundAccountPosition"
	balanceUpdate           = "balanceUpdate"
	executionReport         = "executionReport"
)

type AccountDataWorker struct {
	restClient       *Client
	sugar            *zap.SugaredLogger
	accountInfoStore *common.BinanceAccountInfoStore
}

func NewAccountDataWorker(sugar *zap.SugaredLogger, apiKey, apiSecret string, store *common.BinanceAccountInfoStore) *AccountDataWorker {
	return &AccountDataWorker{
		restClient:       NewBinanceClient(apiKey, apiSecret, sugar),
		sugar:            sugar,
		accountInfoStore: store,
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
		case outboundAccountInfo:
			var ai common.BinanceAccountInfo
			if err = bc.parseAccountInfo(m, &ai); err != nil {
				logger.Errorw("failed to parse account info", "err", err)
				return
			}
			if err := bc.accountInfoStore.SetAccountInfo(&ai); err != nil {
				logger.Errorw("failed to update account info", "error", err)
				return
			}
		case outboundAccountPosition:
			var balance []*common.PayloadBalance
			if bc.parseAccountBalance(m, logger, balance) {
				return
			}
			if err := bc.accountInfoStore.UpdateBalance(balance); err != nil {
				logger.Errorw("failed to update balance info", "error", err)
				return
			}
		case balanceUpdate:
			var balanceUpdate common.BalanceUpdate
			if err := json.Unmarshal(m, &balanceUpdate); err != nil {
				logger.Errorw("failed to unmarshal balanceUpdate", "error", err)
				return
			}
			if err := bc.accountInfoStore.UpdateBalanceDelta(&balanceUpdate); err != nil {
				logger.Errorw("failed to update account balance delta", "error", err)
				return
			}
		case executionReport:
			// TODO: handle later
		}
	}
}

func (bc *AccountDataWorker) parseAccountBalance(m []byte, logger *zap.SugaredLogger, balance []*common.PayloadBalance) bool {
	balanceByte, _, _, err := jsonparser.Get(m, "B")
	if err != nil {
		logger.Errorw("failed to lookup balance", "err", err)
		return true
	}
	if err := json.Unmarshal(balanceByte, &balance); err != nil {
		logger.Errorw("failed to parse balance data", "err", err)
		return true
	}
	return false
}

func (bc *AccountDataWorker) parseAccountInfo(data []byte, accountInfo *common.BinanceAccountInfo) error {
	var err error
	accountInfo.MakerCommission, err = jsonparser.GetInt(data, "m")
	if err != nil {
		return err
	}
	accountInfo.TakerCommission, err = jsonparser.GetInt(data, "t")
	if err != nil {
		return err
	}
	accountInfo.BuyerCommission, err = jsonparser.GetInt(data, "b")
	if err != nil {
		return err
	}
	accountInfo.SellerCommission, err = jsonparser.GetInt(data, "s")
	if err != nil {
		return err
	}
	accountInfo.CanTrade, err = jsonparser.GetBoolean(data, "T")
	if err != nil {
		return err
	}
	accountInfo.CanWithdraw, err = jsonparser.GetBoolean(data, "W")
	if err != nil {
		return err
	}
	accountInfo.CanDeposit, err = jsonparser.GetBoolean(data, "D")
	if err != nil {
		return err
	}
	updateTime, err := jsonparser.GetInt(data, "u")
	if err != nil {
		return err
	}
	accountInfo.UpdateTime = uint64(updateTime)
	accountInfo.AccountType = "SPOT"           // currently we only use this account type
	accountInfo.Permissions = []string{"SPOT"} // this is default permisson
	balanceByte, _, _, err := jsonparser.Get(data, "B")
	if err != nil {
		return err
	}
	var balance []common.PayloadBalance
	if err := json.Unmarshal(balanceByte, &balance); err != nil {
		return err
	}
	for _, tokenBalance := range balance {
		accountInfo.Balances = append(accountInfo.Balances, common.Balance{
			Asset:  tokenBalance.Asset,
			Free:   tokenBalance.Free,
			Locked: tokenBalance.Lock,
		})
	}
	return nil
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
	tm := time.NewTimer(time.Second)
	tm.Stop()
	for {
		_, m, err := wsConn.ReadMessage()
		if err != nil {
			logger.Errorw("read message error", "err", err)
			return err
		}
		log.Printf("%s \n", m)
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
	bc.sugar.Info("listen key ", listenKey)
	// init account info
	binanceAccountInfo, err := bc.restClient.GetAccountInfo()
	if err != nil {
		bc.sugar.Errorw("failed to init account info", "error", err)
		return "", err
	}
	_ = bc.accountInfoStore.SetAccountInfo(&binanceAccountInfo)
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
			bc.sugar.Errorw("subscribe data stream broken, retry after seconds")
			updater.Stop()
			time.Sleep(time.Second * 5)
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
