// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com/).
// Copyright 2018-present StackState

package validate

import (
	"github.com/StackVista/stackstate-agent/pkg/config"
	hostnameValidate "github.com/StackVista/stackstate-agent/pkg/util/hostname/validate"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// ValidHostname determines whether the passed string is a valid hostname.
// In case it's not, the returned error contains the details of the failure.
func ValidHostname(hostname string) error {
	// [sts] If hostname validation is disabled just return nil
	skipHostnameValidation := config.Datadog.GetBool("skip_hostname_validation")
	if skipHostnameValidation {
		log.Debugf("Hostname validation is disabled, accepting %s as a valid hostname", hostname)
		return nil
	}

	return hostnameValidate.ValidHostname(hostname)
}
