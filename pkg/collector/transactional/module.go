package transactional

import (
	"context"
	"github.com/DataDog/datadog-agent/comp/core/hostname"
	"github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"
)

// Module defines the fx options for this component.
func Module() fxutil.Module {
	return fxutil.Component(
		fx.Provide(newTransactionalClient))
}

type dependencies struct {
	fx.In
	Log      log.Component
	hostname hostname.Component

	Params Params
}

type provides struct {
	fx.Out
	TransactionalClient Component
}

func newTransactionalClient(deps dependencies) (provides, error) {
	hname, err := deps.hostname.Get(context.TODO())
	if err != nil {
		return provides{}, err
	}

	return provides{
		TransactionalClient: NewComponent(deps.Params, hname),
	}, nil
}
