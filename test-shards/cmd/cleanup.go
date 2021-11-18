package cmd

import (
	"fmt"
	"log"
	"os"
	"test-shards/driver"
	"test-shards/step"

	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
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
		prepare := step.Prepare(create)
		cleanup := step.Cleanup(prepare)

		vars := map[string]interface{}{
			"rootDir":         eksReceiverDir,
			"ansibleTasksDir": fmt.Sprintf("%s/library/ansible-tasks", cwd),
		}
		driver.AnsiblePlay(cleanup, vars)
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cleanupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cleanupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
