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
		create := findScenario(args[0]).generateCreateStep(runId)
		destroy := step.Destroy(create)
		doDestroy(destroy, !assumeYes)
	},
}

func doDestroy(destroy *step.DestroyStep, prompt bool) {
	driver.TerraformDestroy(destroy, prompt)
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	destroyCmd.Flags().BoolVarP(&assumeYes, AssumeYesFlag, AssumeYesShortFlag, false, "automatic yes to prompts")
}
