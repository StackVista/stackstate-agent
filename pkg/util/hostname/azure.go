package hostname

import "github.com/StackVista/stackstate-agent/pkg/util/cloudproviders/azure"

func init() {
	RegisterHostnameProvider("azure", azure.GetHostname)
}
