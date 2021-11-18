package cmd

import (
	"fmt"
	"log"
	"os"
	"test-shards/driver"
	"test-shards/step"

	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}
		eksReceiverDir := fmt.Sprintf("%s/shards/eks-receiver", cwd)

		create := step.Create(eksReceiverDir)
		destroy := step.Destroy(create)
		tfVars := map[string]string{
			"k8s_version": "1.21",
			"environment": "tiziano-test-shard", // TODO
		}
		driver.TerraformDestroy(destroy, tfVars, true)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
