package binance

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/KyberNetwork/binance_user_data_stream/lib/caller"
	"go.uber.org/zap"
)

const (
	defaultTimeout = 5 * time.Second
)

type BinanceClient struct {
	httpClient *http.Client
	apiKey     string
	secretKey  string
	sugar      *zap.SugaredLogger
}

func NewBinanceClient(key, secret string) *BinanceClient {
	return &BinanceClient{
		httpClient: &http.Client{Timeout: defaultTimeout},
		apiKey:     key,
		secretKey:  secret,
	}
}

//BinanceListenKey is listen for user data stream
type BinanceListenKey struct {
	ListenKey string `json:"listenKey"`
}

func (bc *BinanceClient) createListenKey() (string, error) {
	var (
		listenKey BinanceListenKey
		logger    = bc.sugar.With("func", caller.GetCurrentFunctionName())
	)
	requestURL := "https://api.binance.com/api/v3/userDataStream"
	req, err := http.NewRequest(http.MethodPost, requestURL, nil)
	if err != nil {
		logger.Errorw("failed to create request to create listen key", "error", err)
	}
	req.Header.Set("X-MBX-APIKEY", bc.apiKey)
	resp, err := bc.httpClient.Do(req)
	switch resp.StatusCode {
	case http.StatusOK:
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Errorw("failed to read response body", "error", err)
			return "", err
		}
		if err := json.Unmarshal(respBody, &listenKey); err != nil {
			logger.Errorw("failed to parse json into struct", "error", err)
		}
	default:
		logger.Errorw("got unexpected status code", "code", resp.StatusCode)
	}
	return listenKey.ListenKey, nil
}
