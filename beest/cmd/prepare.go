package cmd

import (
	"beest/cmd/driver"
	"beest/cmd/step"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(prepareCmd)
}

var prepareCmd = &cobra.Command{
	Use:   "prepare [scenario]",
	Short: "Deploy the bees configured as part of the yard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario := findScenario(args[0])
		cobra.CheckErr(prepare(&driver.AnsibleDeployer{}, scenario))
	},
}

func prepare(deployer driver.Deployer, scenario *Scenario) error {
	create := scenario.generateCreateStep(runId)
	prepare := step.Prepare(create)
	return deployer.Prepare(prepare)
}
