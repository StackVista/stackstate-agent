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

	InstallDepsFlag = "deps"
)

var (
	watchTest     bool
	testSelection string
	installDeps   bool
)

func init() {
	rootCmd.AddCommand(verifyCmd)

	verifyCmd.Flags().BoolVarP(&watchTest, WatchFlag, WatchShortFlag, false, "watch for changes and re-run the tests")
	verifyCmd.Flags().StringVar(&testSelection, TestSelectionFlag, "", "a selection of test names to run")
	verifyCmd.Flags().BoolVar(&installDeps, InstallDepsFlag, false, "install deps before running the tests")
}

var verifyCmd = &cobra.Command{
	Use:   "verify [scenario]",
	Short: "Run the tests against the yard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario := findScenario(args[0])
		cobra.CheckErr(verify(&driver.PyTestVerifier{}, scenario, watchTest, testSelection, installDeps))
	},
}

func verify(verifier driver.Verifier, scenario *Scenario, watch bool, selection string, deps bool) error {
	create := scenario.generateCreateStep(runId)
	prepare := step.Prepare(create)
	verify := step.Verify(prepare, scenario.Test.path(), []string{})
	return verifier.Verify(verify, watch, selection, deps)
}
