package transactionalclientimpl

import (
	"github.com/DataDog/datadog-agent/pkg/config/setup"
	"time"
)

type Params struct {
	maxCapacity                                                             int
	transactionChannelBufferSize                                            int
	tickerInterval, transactionTimeoutDuration, transactionEvictionDuration time.Duration
}

// NewDefaultParams returns the default parameters for the transactional stackstate client
func NewDefaultParams() Params {
	txChannelBufferSize, txTickerInterval, txTimeoutDuration, txEvictionDuration := setup.GetTxManagerConfig()

	return Params{
		maxCapacity:                  setup.GetMaxCapacity(),
		transactionChannelBufferSize: txChannelBufferSize,
		tickerInterval:               txTickerInterval,
		transactionTimeoutDuration:   txTimeoutDuration,
		transactionEvictionDuration:  txEvictionDuration,
	}
}
