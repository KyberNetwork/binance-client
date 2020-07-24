package binance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/KyberNetwork/binance_user_data_stream/lib/caller"
	"go.uber.org/zap"
)

const (
	defaultTimeout = 5 * time.Second
)

// Client to interact with binance api
type Client struct {
	httpClient *http.Client
	apiKey     string
	secretKey  string
	sugar      *zap.SugaredLogger
}

// NewBinanceClient create new client object
func NewBinanceClient(key, secret string, sugar *zap.SugaredLogger) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		apiKey:     key,
		secretKey:  secret,
		sugar:      sugar,
	}
}

//ListenKey is listen for user data stream
type ListenKey struct {
	ListenKey string `json:"listenKey"`
}

// CreateListenKey create a listen key for user data stream
func (bc *Client) CreateListenKey() (string, error) {
	var (
		listenKey ListenKey
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

// KeepListenKeyAlive keep it alive
func (bc *Client) KeepListenKeyAlive() error {
	var (
		logger = bc.sugar.With("func", caller.GetCurrentFunctionName())
	)
	requestURL := "https://api.binance.com/api/v3/userDataStream"
	req, err := http.NewRequest(http.MethodPut, requestURL, nil)
	if err != nil {
		logger.Errorw("failed to create new request for keep listen key alive", "error", err)
		return err
	}
	req.Header.Set("X-MBX-APIKEY", bc.apiKey)
	resp, err := bc.httpClient.Do(req)
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		logger.Errorw("got unexpected status code", "code", resp.StatusCode)
		return fmt.Errorf("failed with status code %d", resp.StatusCode)
	}
}
