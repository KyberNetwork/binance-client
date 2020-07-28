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
	"github.com/pkg/errors"
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
		return "",err
	}
	req.Header.Set("X-MBX-APIKEY", bc.apiKey)

	err = bc.doRequest(req, logger, &listenKey)
	if err != nil {
		return "", err
	}
	return listenKey.ListenKey, nil
}

func (bc *Client) doRequest(req *http.Request, logger *zap.SugaredLogger, data interface{}) error {
	resp, err := bc.httpClient.Do(req)
	if err != nil {
		logger.Errorw("failed to execute the request", "error", err)
		return errors.Wrap(err,"failed to execute the request")
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorw("failed to read response body", "error", err)
		return err
	}
	_ = resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		if data==nil{// if data == nil then caller does not care about response body, consider as success
			return nil
		}
		if err = json.Unmarshal(respBody, data); err != nil {
			logger.Errorw("failed to parse data into struct", "error", err)
			return errors.Wrap(err,"failed to parse data into struct")
		}
	default:
		logger.Errorw("got unexpected status code", "code", resp.StatusCode,"responseBody",string(respBody))
		return fmt.Errorf("got unexpected status code %d, body=%s",resp.StatusCode,string(respBody))
	}
	return nil
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
	return bc.doRequest(req,logger,nil)
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

	err = bc.doRequest(req,logger,&response)
	return response, err
}
