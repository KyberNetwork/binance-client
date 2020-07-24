package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	ws "github.com/gorilla/websocket"

	"github.com/KyberNetwork/binance_user_data_stream/common"
	"github.com/KyberNetwork/binance_user_data_stream/lib/caller"
)

const (
	outboundAccountInfo     = "outboundAccountInfo"
	outboundAccountPosition = "outboundAccountPosition"
	balanceUpdate           = "balanceUpdate"
	executionReport         = "executionReport"
)

func (bc *Client) processMessages(messages chan *common.UserDataStreamPayload) {
	var (
		logger = bc.sugar.With("func", caller.GetCurrentFunctionName())
	)
	for m := range messages {
		log.Printf("%s", m)
		switch m.EventType {
		case outboundAccountInfo:
			payload := common.OutBoundAccountInfo{}
			if err := json.Unmarshal(m.Payload, &payload); err != nil {
				logger.Errorw("failed to unmarshal outbound account info", "error", err)
				return
			}
			//TODO:
		case outboundAccountPosition:
			payload := common.OutboundAccountPosition{}
			if err := json.Unmarshal(m.Payload, &payload); err != nil {
				logger.Errorw("failed to unmarshal outbound account position", "error", err)
				return
			}
		case balanceUpdate:
			payload := common.BalanceUpdate{}
			if err := json.Unmarshal(m.Payload, &payload); err != nil {
				logger.Errorw("failed to unmarshal balance update", "error", err)
				return
			}
		case executionReport:
			payload := common.ExecutionReport{}
			if err := json.Unmarshal(m.Payload, &payload); err != nil {
				logger.Errorw("failed to unmarshal balance update", "error", err)
				return
			}
		}
	}
}

// SubscribeDataStream subscribe to a data stream
func (bc *Client) SubscribeDataStream(listenKey string, messages chan<- *common.UserDataStreamPayload) error {
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
		message := common.UserDataStreamPayload{}
		err := wsConn.ReadJSON(&message)
		if err != nil {
			logger.Errorw("read message error", "err", err)
			return err
		}
		tm.Reset(time.Second)
		log.Printf("%s", message)
		select {
		case messages <- &message:
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
	messages := make(chan *common.UserDataStreamPayload, 256)
	go bc.processMessages(messages)
	return bc.SubscribeDataStream(listenKey, messages)
}
