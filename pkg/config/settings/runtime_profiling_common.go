package settings

import (
	"errors"

	"github.com/StackVista/stackstate-agent/pkg/util/profiling"
)

func checkProfilingNeedsRestart(old, new int) error {
	if old == 0 && new != 0 && profiling.IsRunning() {
		return errors.New("go runtime setting applied; manually restart internal profiling to capture profile data in the app")
	}
	return nil
}
