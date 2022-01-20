package cmd

import (
	"beest/cmd/driver"
	"beest/cmd/step"
	"github.com/spf13/cobra"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare [scenario]",
	Short: "Deploy the bees configured as part of the yard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario = choseScenario(args[0])

		create := step.Create(scenario.Yard.path(), scenario.mergeVars(commonVariables()))
		prepare := step.Prepare(create)
		doPrepare(prepare)
	},
}

func doPrepare(prepare *step.PrepareStep) {
	driver.AnsiblePlay(prepare)
}

func init() {
	rootCmd.AddCommand(prepareCmd)
}
