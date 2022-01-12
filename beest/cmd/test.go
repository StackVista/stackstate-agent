package cmd

import (
	"beest/cmd/step"
	"github.com/spf13/cobra"
	"log"
)

const (
	NoDestroyFlag = "no-destroy"
	ResetFlag     = "reset"
)

var (
	noDestroy bool
	reset     bool
)

var testCmd = &cobra.Command{
	Use:   "test [scenario]",
	Short: "Execute all the steps in sequence",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario = choseScenario(args[0])

		create := step.Create(scenario.Yard.path(), scenario.mergeVars(commonVariables()))
		doCreate(create, false)
		prepare := step.Prepare(create)
		cleanup := step.Cleanup(prepare)
		if reset {
			doCleanup(cleanup)
		}
		doPrepare(prepare)
		verify := step.Verify(prepare, scenario.Test.path(), []string{})
		testError := doVerify(verify, false, "")
		defer cobra.CheckErr(testError)

		doCleanup(cleanup)
		if !noDestroy {
			destroy := step.Destroy(create)
			doDestroy(destroy, false)
		} else {
			log.Println("The yard won't be destroyed")
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().BoolVar(&noDestroy, NoDestroyFlag, false, "do not destroy the yard")
	testCmd.Flags().BoolVar(&reset, ResetFlag, false, "execute a cleanup before prepare if tests run already")
}
