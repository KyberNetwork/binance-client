package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/KyberNetwork/binance_user_data_stream/common"
	"github.com/KyberNetwork/binance_user_data_stream/lib/caller"
)

const (
	defaultTimeout = 5 * time.Second
	apiBaseURL     = "https://api.binance.com"
)

// Client to interact with binance api
type Client struct {
	httpClient       *http.Client
	apiKey           string
	secretKey        string
	sugar            *zap.SugaredLogger
	accountInfoStore *common.AccountInfoStore
}

// NewBinanceClient create new client object
func NewBinanceClient(key, secret string, sugar *zap.SugaredLogger, accountInfoStore *common.AccountInfoStore) *Client {
	return &Client{
		httpClient:       &http.Client{Timeout: defaultTimeout},
		apiKey:           key,
		secretKey:        secret,
		sugar:            sugar,
		accountInfoStore: accountInfoStore,
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
	requestURL := fmt.Sprintf("%s/api/v3/userDataStream", apiBaseURL)
	req, err := http.NewRequest(http.MethodPost, requestURL, nil)
	if err != nil {
		logger.Errorw("failed to create request to create listen key", "error", err)
	}
	req.Header.Set("X-MBX-APIKEY", bc.apiKey)
	resp, err := bc.httpClient.Do(req)
	if err != nil {
		logger.Errorw("failed to do the request to create listen key", "error", err)
		return "", err
	}
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
	requestURL := fmt.Sprintf("%s/api/v3/userDataStream", apiBaseURL)
	req, err := http.NewRequest(http.MethodPut, requestURL, nil)
	if err != nil {
		logger.Errorw("failed to create new request for keep listen key alive", "error", err)
		return err
	}
	req.Header.Set("X-MBX-APIKEY", bc.apiKey)
	resp, err := bc.httpClient.Do(req)
	if err != nil {
		logger.Errorw("failed to do the request to keep the key alive", "error", err)
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		logger.Errorw("got unexpected status code", "code", resp.StatusCode)
		return fmt.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// Sign the request
func (bc *Client) Sign(msg string) string {
	mac := hmac.New(sha256.New, []byte(bc.secretKey))
	if _, err := mac.Write([]byte(msg)); err != nil {
		panic(err) // should never happen
	}
	result := ethereum.Bytes2Hex(mac.Sum(nil))
	return result
}

func getTimepoint() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
}

// GetAccountInfo return account info
func (bc *Client) GetAccountInfo() (common.BinanceAccountInfo, error) {
	var (
		logger   = bc.sugar.With("func", caller.GetCurrentFunctionName())
		response common.BinanceAccountInfo
	)
	requestURL := fmt.Sprintf("%s/api/v3/account", apiBaseURL)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		logger.Errorw("failed to create new request for get account info", "error", err)
	}

	// sign the request
	req.Header.Set("X-MBX-APIKEY", bc.apiKey)
	q := req.URL.Query()
	sig := url.Values{}
	q.Set("timestamp", fmt.Sprintf("%d", getTimepoint()))
	q.Set("recvWindow", "5000")
	sig.Set("signature", bc.Sign(q.Encode()))
	req.URL.RawQuery = q.Encode() + "&" + sig.Encode()

	resp, err := bc.httpClient.Do(req)
	if err != nil {
		logger.Errorw("failed to do get account info request", "error", err)
	}
	switch resp.StatusCode {
	case http.StatusOK:
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Errorw("failed to read response body", "error", err)
			return response, err
		}
		if err := json.Unmarshal(respBody, &response); err != nil {
			return response, err
		}
	default:
		logger.Errorw("got unexpected status code", "code", resp.StatusCode)
		return response, fmt.Errorf("failed with status code %d", resp.StatusCode)
	}
	return response, nil
}
