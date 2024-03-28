package teststate

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/rtloader/test/helpers"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	err := setUp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up tests: %v", err)
		os.Exit(-1)
	}

	os.Exit(m.Run())
}

func TestSetState(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	stateKey := "key"
	stateBody := "state body"

	out, err := run(fmt.Sprintf(`state.set_state(None, "checkid", "%s", "%s")`, stateKey, stateBody))

	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}
	if checkID != "checkid" {
		t.Fatalf("Unexpected check id value: %s", checkID)
	}
	if stateStorage[stateKey] != stateBody {
		t.Fatalf("Unexpected saved state value: %s", stateStorage[stateKey])
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

func TestGetState(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	stateKey := "key"
	stateBody := "{}"

	// Insert temp value into the state storage allow the get_state function to retrieve it
	stateStorage[stateKey] = stateBody

	// state.set_state(None, "checkid", "%s", "%s");
	out, err := run(fmt.Sprintf(`state.get_state(None, "checkid", "%s")`, stateKey))

	if err != nil {
		t.Fatal(err)
	}

	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}

	if checkID != "checkid" {
		t.Fatalf("Unexpected check id value: %s", checkID)
	}

	if lastRetrievedState != stateBody {
		t.Fatalf("Unexpected retieved state value: %s", lastRetrievedState)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}
