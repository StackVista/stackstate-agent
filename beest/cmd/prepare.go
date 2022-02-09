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
		create := findScenario(args[0]).generateCreateStep(runId)
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
