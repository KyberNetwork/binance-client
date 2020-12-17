package binance

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// CoinStateWorker monitor binance coin state, we use this to check asset withdrawable before execute withdraw
type CoinStateWorker struct {
	bc       *Client
	l        *zap.SugaredLogger
	snapshot AllCoinInfo
	allState map[string]CoinInfo
	lock     sync.Mutex
}

func (s *CoinStateWorker) update() {
	i, err := s.bc.AllCoinInfo()
	if err != nil {
		s.l.Errorw("get all coin info failed", "err", err)
		return
	}
	as := i.ToMap()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.snapshot = i
	s.allState = as
}

func (s *CoinStateWorker) run() {
	for range time.NewTicker(time.Minute).C {
		s.update()
	}
}

func NewCoinStateWorker(l *zap.SugaredLogger, c *Client) *CoinStateWorker {
	res := &CoinStateWorker{
		bc: c,
		l:  l,
	}
	res.update() // pre-fetch first time
	go res.run()
	return res
}

func (s *CoinStateWorker) AllInfo() AllCoinInfo {
	s.lock.Lock()
	res := s.snapshot
	s.lock.Unlock()
	return res
}

func (s *CoinStateWorker) GetOneInfo(asset string) *CoinInfo {
	s.lock.Lock()
	ci, ok := s.allState[asset]
	s.lock.Unlock()
	if ok {
		return &ci
	}
	return nil
}
