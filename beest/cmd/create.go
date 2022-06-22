package cmd

import (
	"beest/cmd/driver"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().BoolVarP(&assumeYes, AssumeYesFlag, AssumeYesShortFlag, false, "automatic yes to prompts")
}

var createCmd = &cobra.Command{
	Use:   "create [scenario]",
	Short: "Provision the yard used by a certain scenario",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario := findScenario(args[0])
		cobra.CheckErr(create(&driver.TerraformProvisioner{}, scenario, !assumeYes))
	},
}

func create(provisioner driver.Provisioner, scenario *Scenario, prompt bool) error {
	create := scenario.generateCreateStep(runId)
	return provisioner.Create(create, prompt)
}
