package cmd

import (
	"beest/cmd/driver"
	"beest/cmd/step"
	"github.com/spf13/cobra"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare [scenario]",
	Short: "Deploy the integrations configured as part of the yard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario = choseScenario(args[0])

		create := step.Create(scenario.Yard.path())
		prepare := step.Prepare(create)
		doPrepare(prepare)
	},
}

func doPrepare(prepare *step.PrepareStep) {
	driver.AnsiblePlay(prepare, map[string]interface{}{})
}

func init() {
	rootCmd.AddCommand(prepareCmd)
}
