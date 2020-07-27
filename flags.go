package binance

import (
	"github.com/KyberNetwork/binance_user_data_stream/common"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

const (
	binanceKeyFlag    = "binance-key"
	binanceSecretFlag = "binance-secret"
)

//NewBinanceFlags return flags for binance client
func NewBinanceFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   binanceKeyFlag,
			Usage:  "binance key",
			EnvVar: "BINANCE_KEY",
		},
		cli.StringFlag{
			Name:   binanceSecretFlag,
			Usage:  "binance secret",
			EnvVar: "BINANCE_SECRET",
		},
	}
}

// NewBinanceClientFromContext create binance client from flags
func NewBinanceClientFromContext(c *cli.Context, sugar *zap.SugaredLogger, accountInfoStore *common.AccountInfoStore) *Client {

	// TODO: add validation
	binanceKey := c.String(binanceKeyFlag)
	binanceSecret := c.String(binanceSecretFlag)

	return NewBinanceClient(binanceKey, binanceSecret, sugar, accountInfoStore)
}
