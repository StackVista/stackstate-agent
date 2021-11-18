package cmd

import (
	"beest/cmd/driver"
	"beest/cmd/step"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [scenario]",
	Short: "Provision the yard used by a certain scenario",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario = choseScenario(args[0])

		create := step.Create(scenario.Yard.path())
		doCreate(create, true)
	},
}

func doCreate(create *step.CreationStep, prompt bool) {
	driver.TerraformApply(create, scenario.mergeVars(commonVariables()), prompt)
}

func init() {
	rootCmd.AddCommand(createCmd)
}
