package cmd

import (
	"beest/cmd/driver"
	"beest/cmd/step"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy [scenario]",
	Short: "Destroy all resources associated with the yard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario = choseScenario(args[0])

		create := step.Create(scenario.Yard.path())
		destroy := step.Destroy(create)
		doDestroy(destroy, !assumeYes)
	},
}

func doDestroy(destroy *step.DestroyStep, prompt bool) {
	driver.TerraformDestroy(destroy, scenario.mergeVars(commonVariables()), prompt)
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	destroyCmd.Flags().BoolVarP(&assumeYes, AssumeYesFlag, AssumeYesShortFlag, false, "automatic yes to prompts")
}
