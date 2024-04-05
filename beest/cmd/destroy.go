package cmd

import (
	"beest/cmd/driver"
	"beest/cmd/step"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(destroyCmd)

	destroyCmd.Flags().BoolVarP(&assumeYes, AssumeYesFlag, AssumeYesShortFlag, false, "automatic yes to prompts")
}

var destroyCmd = &cobra.Command{
	Use:   "destroy [scenario]",
	Short: "Destroy all resources associated with the yard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario := findScenario(args[0])
		cobra.CheckErr(destroy(&driver.TerraformProvisioner{}, scenario, !assumeYes))
	},
}

func destroy(provisioner driver.Provisioner, scenario *Scenario, prompt bool) error {
	create := scenario.generateCreateStep(runId)
	destroy := step.Destroy(create)
	return provisioner.Destroy(destroy, prompt)
}
