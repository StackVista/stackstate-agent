package cmd

import (
	"beest/cmd/driver"
	"beest/cmd/step"
	"github.com/spf13/cobra"
)

var (
	verifyExclusions []string
	verifyInclusions []string
)

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().StringArrayVarP(&verifyExclusions, ExclusionsFlag, ExclusionsShortFlag, []string{}, "exclude certain bees")
	cleanupCmd.Flags().StringArrayVarP(&verifyInclusions, InclusionsFlag, InclusionsShortFlag, []string{}, "cleanup only the specified yard role")
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup [scenario]",
	Short: "Undeploy all the configured bees",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenario := findScenario(args[0])
		cobra.CheckErr(cleanup(&driver.AnsibleDeployer{}, scenario))
	},
}

func cleanup(deployer driver.Deployer, scenario *Scenario) error {
	create := scenario.generateCreateStep(runId)
	prepare := step.Prepare(create)
	cleanup := step.Cleanup(prepare)
	return deployer.Cleanup(cleanup, verifyExclusions, verifyInclusions)
}
