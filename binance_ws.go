package binance

import (
	"fmt"
	"log"
	"time"

	"github.com/buger/jsonparser"
	ws "github.com/gorilla/websocket"

	"github.com/KyberNetwork/binance_user_data_stream/lib/caller"
)

const (
	outboundAccountInfo     = "outboundAccountInfo"
	outboundAccountPosition = "outboundAccountPosition"
	balanceUpdate           = "balanceUpdate"
	executionReport         = "executionReport"
)

func (bc *Client) processMessages(messages chan []byte) {
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
			if err := bc.accountInfoStore.UpdateAccountInfo(m); err != nil {
				logger.Errorw("failed to update account info", "error", err)
				return
			}
		case outboundAccountPosition:
			if err := bc.accountInfoStore.UpdateBalance(m); err != nil {
				logger.Errorw("failed to update balance info", "error", err)
				return
			}
		case balanceUpdate:
			if err := bc.accountInfoStore.UpdateBalanceDelta(m); err != nil {
				logger.Errorw("failed to update account balance delta", "error", err)
				return
			}
		case executionReport:
			//TODO: handle later
		}
	}
}

// SubscribeDataStream subscribe to a data stream
func (bc *Client) SubscribeDataStream(listenKey string, messages chan<- []byte) error {
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
	defer wsConn.Close()
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

// Run the websocket
func (bc *Client) Run(listenKey string) error {
	messages := make(chan []byte, 256)
	go bc.processMessages(messages)
	return bc.SubscribeDataStream(listenKey, messages)
}
