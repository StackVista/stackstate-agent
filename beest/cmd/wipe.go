package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

/*
TODO:
- Use colon `:` as a delimiter between run_id and scenario_name
  - throw error if used
- Scan through s3 bucket (get_env("BEEST_S3_BUCKET"))
  - parse
s3 key: STAC-15754-zerouco:dockerd-eks/stackstate+gitlabci_agentv2/tf.tfstate

// STAC-15754-zerouco -> $RUN_ID
// dockerd-eks -> scenario arg
// stackstate+gitlabci -> $quay_user
// 2 -> $MAJOR_VERSION
*/

// wipeCmd represents the wipe command
var wipeCmd = &cobra.Command{
	Use:   "wipe",
	Short: "Wipe beest yards older than ",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("wipe called")
	},
}

func init() {
	rootCmd.AddCommand(wipeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wipeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wipeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
