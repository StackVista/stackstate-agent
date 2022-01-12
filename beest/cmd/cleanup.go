package cmd

import (
	"beest/cmd/driver"
	"beest/cmd/step"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup [scenario]",
	Short: "Uninstall all the configured integrations",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario = choseScenario(args[0])

		create := step.Create(scenario.Yard.path(), scenario.mergeVars(commonVariables()))
		prepare := step.Prepare(create)
		cleanup := step.Cleanup(prepare)
		doCleanup(cleanup)
	},
}

func doCleanup(cleanup *step.CleanupStep) {
	driver.AnsiblePlay(cleanup)
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
}
