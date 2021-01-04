package binance

import (
	"fmt"
	"sync"
	"time"
)

type accountID = string

type MarginAccountInfo struct {
	accounts           map[accountID]*BContext
	accountDetails     map[accountID]*CrossMarginAccountDetails
	accountsLastUpdate map[accountID]time.Time
	lock               sync.Mutex
}

func NewMarginAccountInfo(accounts map[accountID]*BContext) *MarginAccountInfo {
	return &MarginAccountInfo{
		accounts:           accounts,
		accountsLastUpdate: make(map[accountID]time.Time),
		accountDetails:     make(map[accountID]*CrossMarginAccountDetails),
	}
}

func (m *MarginAccountInfo) UpdateAccount(id accountID) (*CrossMarginAccountDetails, error) {
	acc := m.accounts[id]
	if acc == nil {
		return nil, fmt.Errorf("account not exists %s", id)
	}
	ai, _, err := acc.RestClient.GetCrossMarginAccountDetails()
	if err != nil {
		return nil, err
	}
	m.lock.Lock()
	m.accountDetails[id] = &ai
	m.accountsLastUpdate[id] = time.Now()
	m.lock.Unlock()
	return &ai, nil
}

// GetAccountInfo get margin account detail for specify account
func (m *MarginAccountInfo) GetAccountInfo(id accountID) (*CrossMarginAccountDetails, error) {
	m.lock.Lock()
	update := m.accountsLastUpdate[id]
	ai, ok := m.accountDetails[id]
	m.lock.Unlock()
	const cachedValidDuration = time.Minute
	if ok && time.Since(update) <= cachedValidDuration { // if found a valid local cache copy, return it
		return ai, nil
	}
	ai, err := m.UpdateAccount(id) // update new account info
	if err != nil {
		return nil, err
	}
	return ai, nil
}
