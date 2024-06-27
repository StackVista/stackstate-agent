package handler

import (
	"github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/pkg/batcher"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"
)

// Bundle give a StackState-specific bundle with batcher, transactional and check handler.
func Bundle() fxutil.BundleOptions {
	return fxutil.Bundle(
		fx.Supply(batcher.NewDefaultParams()),
		batcher.Module(),
		fx.Supply(transactional.NewDefaultParams()),
		transactional.Module(),
		Module())
}

// Module defines the fx options for this component.
func Module() fxutil.Module {
	return fxutil.Component(
		fx.Provide(newCheckManager))
}

type dependencies struct {
	fx.In
	Log                 log.Component
	TransactionalClient transactional.Component
	Batcher             batcher.Component
}

func newCheckManager(deps dependencies) (CheckManager, error) {
	return NewCheckManager(deps.Batcher, deps.TransactionalClient.GetBatcher(), deps.TransactionalClient.GetManager()), nil
}
