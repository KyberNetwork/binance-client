package binance

import (
	"github.com/KyberNetwork/cex_account_data/lib/ocache"
	"github.com/KyberNetwork/cex_account_data/lib/orderlist"
)

// BContext contain object to operate with binance for one account context
type BContext struct {
	AccountInfoStore *BAccountInfoStore
	RestClient       *Client
	WSOrderTracker   *orderlist.OrderList
	CompletedOrders  *ocache.OCache
	OrderTrackMillis int64
	MainClient       *Client
}
