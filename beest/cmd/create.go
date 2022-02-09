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
		create := findScenario(args[0]).generateCreateStep(runId)
		doCreate(create, !assumeYes)
	},
}

func doCreate(create *step.CreationStep, prompt bool) {
	driver.TerraformApply(create, prompt)
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().BoolVarP(&assumeYes, AssumeYesFlag, AssumeYesShortFlag, false, "automatic yes to prompts")
}
