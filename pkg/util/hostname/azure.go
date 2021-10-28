// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.

package hostname

import (
	"github.com/StackVista/stackstate-agent/pkg/util/azure"
)

func init() {
	RegisterHostnameProvider("azure", azure.HostnameProvider)
}
