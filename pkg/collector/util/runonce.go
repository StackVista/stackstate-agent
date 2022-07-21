package util

import "go.uber.org/atomic"

type RunOnce interface {
	RunOnce(func())
}

type runOnce struct {
	wasUsed *atomic.Bool
}

func (r *runOnce) RunOnce(f func()) {
	if r.wasUsed.CAS(false, true) {
		f()
	}
}

func NewRunOnce() RunOnce {
	return &runOnce{atomic.NewBool(false)}
}
