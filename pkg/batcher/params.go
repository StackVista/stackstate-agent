package batcher

import "github.com/DataDog/datadog-agent/pkg/config/setup"

type Params struct {
	maxCapacity int
}

func NewDefaultParams() Params {
	return Params{
		maxCapacity: setup.GetMaxCapacity(),
	}
}
