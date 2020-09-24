package binance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestParseAccountBalance(t *testing.T) {
	data := []byte(`{"e":"outboundAccountPosition","E":1600749480712,"u":1600749480711,"B":[{"a":"BNB","f":"1.60268308","l":"0.00000000"},{"a":"USDT","f":"1000.24006746","l":"0.00000000"},{"a":"UNI","f":"1.54197400","l":"75.34000000"}]}`)
	sugar := zap.L().Sugar()
	accountDataWorker := NewAccountDataWorker(sugar, nil) // random account id
	balance := []*PayloadBalance{}
	err := accountDataWorker.parseAccountBalance(data, sugar, &balance)
	require.NoError(t, err)
	assert.Equal(t, balance[0].Asset, "BNB")
	assert.Equal(t, balance[0].Free, "1.60268308")
	assert.Equal(t, balance[0].Lock, "0.00000000")
}
