package stackstate

import (
	"github.com/DataDog/datadog-agent/comp/stackstate/batcher/batcherimpl"
	"github.com/DataDog/datadog-agent/comp/stackstate/checkmanager/checkmanagerimpl"
	"github.com/DataDog/datadog-agent/comp/stackstate/transactionalclient/transactionalclientimpl"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"
)

func Bundle() fxutil.BundleOptions {
	return fxutil.Bundle(
		fx.Supply(batcherimpl.NewDefaultParams()),
		batcherimpl.Module(),
		fx.Supply(transactionalclientimpl.NewDefaultParams()),
		transactionalclientimpl.Module(),
		checkmanagerimpl.Module())
}
