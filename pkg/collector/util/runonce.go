package util

import "go.uber.org/atomic"

// RunOnce ensures that given callback is being executed only at first time
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

// NewRunOnce create RunOnce structures that ensure single execution of desired piece of code
func NewRunOnce() RunOnce {
	return &runOnce{atomic.NewBool(false)}
}
