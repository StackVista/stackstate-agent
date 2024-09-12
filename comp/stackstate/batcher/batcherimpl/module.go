package batcherimpl

import (
	"context"
	"github.com/DataDog/datadog-agent/comp/aggregator/demultiplexer"
	"github.com/DataDog/datadog-agent/comp/core/hostname"
	"github.com/DataDog/datadog-agent/pkg/batcher"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"
)

// Module defines the fx options for this component.
func Module() fxutil.Module {
	return fxutil.Component(
		fx.Provide(newAsynchronousBatcher))
}

type dependencies struct {
	fx.In
	Comp  demultiplexer.Component
	HName hostname.Component

	Params Params
}

type provides struct {
	fx.Out
	Batcher batcher.Component
}

func newAsynchronousBatcher(deps dependencies) (provides, error) {
	hname, err := deps.HName.Get(context.TODO())
	if err != nil {
		return provides{}, err
	}

	return provides{
		Batcher: batcher.MakeAsynchronousBatcher(deps.Comp.Serializer(), hname, deps.Params.maxCapacity),
	}, nil
}
