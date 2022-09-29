package cmd

import (
	"beest/cmd/driver"
	"beest/cmd/step"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

const (
	NoDestroyFlag = "no-destroy"
	ResetFlag     = "reset"
)

var (
	noDestroy bool
	reset     bool
)

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().BoolVar(&noDestroy, NoDestroyFlag, false, "do not destroy the yard")
	testCmd.Flags().BoolVar(&reset, ResetFlag, false, "execute a cleanup before prepare if tests run already")
}

var testCmd = &cobra.Command{
	Use:   "test [scenario]",
	Short: "Execute all the steps in sequence",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario := findScenario(args[0])
		provisioner := &driver.TerraformProvisioner{}
		deployer := &driver.AnsibleDeployer{}
		verifier := &driver.PyTestVerifier{}
		var errored bool
		for _, e := range test(provisioner, deployer, verifier, scenario, reset, noDestroy) {
			fmt.Println(e.Error())
			errored = true
		}
		if errored {
			os.Exit(1)
		}
	},
}

func test(provisioner driver.Provisioner, deployer driver.Deployer, verifier driver.Verifier, scenario *Scenario, reset, noDestroy bool) []error {
	create := scenario.generateCreateStep(runId)
	prepare := step.Prepare(create)
	cleanup := step.Cleanup(prepare)
	verify := step.Verify(prepare, []string{})
	destroy := step.Destroy(create)

	if err := provisioner.Create(create, false); err != nil {
		return []error{err}
	}

	var cleanupError error
	if reset {
		// if cleanup fails during reset, stop
		if err := deployer.Cleanup(cleanup, []string{}); err != nil {
			return []error{err}
		}
	}

	prepareError := deployer.Prepare(prepare, []string{})
	var verifyError error
	if prepareError == nil {
		// if prepare did not fail, verify
		verifyError = verifier.Verify(verify, false, "")
	} else {
		// if prepare failed, skip verify, but continue test sequence
		log.Printf("Not running verify step because prepare failed: %s\n", prepareError)
	}

	cleanupError = deployer.Cleanup(cleanup, []string{})
	var destroyError error
	if !noDestroy {
		destroyError = provisioner.Destroy(destroy, false)
	} else {
		log.Println("The yard won't be destroyed")
	}

	var errors []error
	errors = appendIfNotNil(prepareError, errors)
	errors = appendIfNotNil(verifyError, errors)
	errors = appendIfNotNil(cleanupError, errors)
	errors = appendIfNotNil(destroyError, errors)
	return errors
}

func appendIfNotNil(prepareError error, errors []error) []error {
	if prepareError != nil {
		errors = append(errors, prepareError)
	}
	return errors
}
