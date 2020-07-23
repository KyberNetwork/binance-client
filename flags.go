package binance

import "github.com/urfave/cli"

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

func NewBinanceClientFromContext(c *cli.Context) *BinanceClient {

	// TODO: add validation
	binanceKey := c.String(binanceKeyFlag)
	binanceSecret := c.String(binanceSecretFlag)

	return NewBinanceClient(binanceKey, binanceSecret)
}
