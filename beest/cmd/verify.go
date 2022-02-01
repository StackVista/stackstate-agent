package cmd

import (
	"beest/cmd/driver"
	"beest/cmd/step"
	"github.com/spf13/cobra"
)

const (
	WatchFlag      = "watch"
	WatchShortFlag = "w"

	TestSelectionFlag = "select"
)

var (
	watchTest     bool
	testSelection string
)

var verifyCmd = &cobra.Command{
	Use:   "verify [scenario]",
	Short: "Run the tests against the yard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario = choseScenario(args[0])

		create := step.Create(scenario.Yard.path(), scenario.mergeVars(commonVariables()))
		prepare := step.Prepare(create)
		verify := step.Verify(prepare, scenario.Test.path(), []string{})
		testError := doVerify(verify, watchTest, testSelection)
		cobra.CheckErr(testError)
	},
}

func doVerify(verify *step.VerificationStep, watch bool, selection string) error {
	return driver.PyTestRun(verify, watch, selection)
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	verifyCmd.Flags().BoolVarP(&watchTest, WatchFlag, WatchShortFlag, false, "watch for changes and re-run the tests")
	verifyCmd.Flags().StringVar(&testSelection, TestSelectionFlag, "", "a selection of test names to run")
}
