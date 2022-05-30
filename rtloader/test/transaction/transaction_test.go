package testtransaction

import (
	"fmt"
	"os"
	"testing"

	"github.com/StackVista/stackstate-agent/rtloader/test/helpers"
)

func TestMain(m *testing.M) {
	err := setUp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up tests: %v", err)
		os.Exit(-1)
	}

	os.Exit(m.Run())
}

func TestStartTransaction(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`transaction.start_transaction(None, "checkid")`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}
	if checkID != "checkid" {
		t.Fatalf("Unexpected check id value: %s", checkID)
	}
	if transactionID != checkID+"-transaction-id" {
		t.Fatalf("Unexpected transaction id value: %s", transactionID)
	}
	if !transactionStarted {
		t.Fatalf("Unexpected transaction stated value: %v", transactionStarted)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

func TestStopTransaction(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`transaction.stop_transaction(None, "checkid")`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}
	if checkID != "checkid" {
		t.Fatalf("Unexpected check id value: %s", checkID)
	}
	if transactionID != "" {
		t.Fatalf("Unexpected transaction id value: %s", transactionID)
	}
	if !transactionCompleted {
		t.Fatalf("Unexpected transaction completed value: %v", transactionCompleted)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}
