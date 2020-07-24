package binance

import "github.com/KyberNetwork/binance_user_data_stream/common"

// AccountInfo for binance account
type AccountInfo interface {
	GetAccountInfo() (common.BinanceAccountInfo, error)
}
