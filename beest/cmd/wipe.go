package cmd

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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
		doWipe()
	},
}

func doWipe() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(os.Getenv("BEEST_S3_BUCKET")),
	})
	if err != nil {
		log.Fatal(err)
	}
	scenarios := loadScenarios().Scenarios

	limit := time.Now().Add(-24 * time.Hour)
	for _, object := range output.Contents {
		if object.LastModified.After(limit) {
			keyString := aws.ToString(object.Key)
			workspace := strings.Split(keyString, "/")[0]
			log.Printf("key=%s size=%d lastModified=%s", keyString, object.Size, object.LastModified)
			log.Printf("workspace=%s", workspace)
			if strings.Contains(workspace, ":") {
				// Normalized RUN_ID
				keyParts := strings.Split(workspace, ":")
				log.Printf("key_id = %s | scenario = %s", keyParts[0], keyParts[1])
			} else {
				// Non-normalized RUN_ID
				for _, scenario := range scenarios {
					if strings.Contains(workspace, scenario.Name) {
						keyIdLength := len(workspace) - len(scenario.Name) - 1
						keyId := workspace[:keyIdLength]
						log.Printf("key_id = %s | scenario = %s", keyId, scenario.Name)
					}
				}
			}
		}
	}
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
