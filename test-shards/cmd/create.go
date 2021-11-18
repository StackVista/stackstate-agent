package cmd

import (
	"fmt"
	"log"
	"os"
	"test-shards/driver"
	"test-shards/step"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
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

		// TODO loop tru variables and ask for input
		//for _, variable := range create.Module.Variables {
		//	log.Println(fmt.Sprintf("variables: %s -> %v", variable.Name, variable.Default))
		//}

		tfVars := map[string]string{
			"k8s_version": "1.21",
			"environment": "tiziano-test-shard", // TODO
		}
		driver.TerraformApply(create, tfVars, true)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
