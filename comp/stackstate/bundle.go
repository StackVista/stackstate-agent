package stackstate

import (
	"github.com/DataDog/datadog-agent/comp/stackstate/batcher/batcherimpl"
	"github.com/DataDog/datadog-agent/comp/stackstate/checkmanager/checkmanagerimpl"
	"github.com/DataDog/datadog-agent/comp/stackstate/transactionalclient/transactionalclientimpl"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
)

func Bundle() fxutil.BundleOptions {
	return fxutil.Bundle(
		batcherimpl.Module(),
		transactionalclientimpl.Module(),
		checkmanagerimpl.Module())
}
