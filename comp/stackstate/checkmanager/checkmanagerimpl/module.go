package checkmanagerimpl

import (
	logdef "github.com/DataDog/datadog-agent/comp/core/log/def"
	"github.com/DataDog/datadog-agent/comp/stackstate/transactionalclient"
	"github.com/DataDog/datadog-agent/pkg/batcher"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"
)

// Module defines the fx options for this component.
func Module() fxutil.Module {
	return fxutil.Component(
		fx.Provide(newCheckManager))
}

type dependencies struct {
	fx.In
	Log                 logdef.Component
	TransactionalClient transactionalclient.Component
	Batcher             batcher.Component
}

func newCheckManager(deps dependencies) (handler.CheckManager, error) {
	return handler.NewCheckManager(deps.Batcher, deps.TransactionalClient.GetBatcher(), deps.TransactionalClient.GetManager()), nil
}
