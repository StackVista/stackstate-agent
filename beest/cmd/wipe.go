package cmd

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pkg/errors"
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

var (
	emptyContext  = context.TODO()
	s3Bucket      = os.Getenv("BEEST_S3_BUCKET")
	dynamodbTable = os.Getenv("BEEST_DYNAMODB_TABLE")
	scenarios     = loadScenarios().Scenarios
)

// wipeCmd represents the wipe command
var wipeCmd = &cobra.Command{
	Use:   "wipe",
	Short: "Wipe beest yards older than 24 hours",
	Run: func(cmd *cobra.Command, args []string) {
		assumeYes = true
		doWipe()
	},
}

type Workspace struct {
	keyId        string
	scenario     string
	username     string
	majorVersion string
}

func doWipe2() {
	runDestroyCmd("containerd-eks", "STAC-15758-fix-pro")
	//"STAC-15758-fix-pro"
	//"containerd-eks"
}

func doWipe() {
	awsCfg, err := config.LoadDefaultConfig(emptyContext)
	if err != nil {
		log.Fatalf("Could not load default config: %s", err)
	}

	// AWS DynamoDB
	dynamodbClient := dynamodb.NewFromConfig(awsCfg)

	limit := time.Now().Add(-24 * time.Hour)
	for _, object := range getS3Objects(awsCfg) {
		keyString := aws.ToString(object.Key)
		log.Printf("key=%s size=%d lastModified=%s", keyString, object.Size, object.LastModified)
		workspace := strings.Split(keyString, "/")[0]
		if object.LastModified.Before(limit) {
			log.Printf("workspace '%s' more than 24 hours old. Cleaning up...", workspace)
			keyId, scenarioName, err := extractVariables(workspace)
			log.Printf("key_id = %s | scenario = %s", keyId, scenarioName)
			if err != nil {
				log.Println(err)
				continue
			}
			runDestroyCmd(scenarioName, keyId)

			err = deleteDynamoDBItem(dynamodbClient, workspace)
			if err != nil {
				log.Printf("Could not delete DynamoDB item because of error: %s", err.Error())
			}

			// TODO: Destroy Dynamo entry
			// TODO: Destroy s3 object
		} else {
			log.Printf("workspace '%s' was used in the last 24 hours.", workspace)
		}
	}
}

type Item struct {
	LockID string
	Digest string
}

func deleteDynamoDBItem(dbClient *dynamodb.Client, workspace string) error {
	log.Printf("Deleting DynamoDB Item from workspace '%s'", workspace)
	expr, err := expression.NewBuilder().WithFilter(
		expression.Contains(expression.Name("LockID"), workspace),
	).Build()
	if err != nil {
		return errors.Errorf("Could not build DynamoDB expression: %s", err.Error())
	}

	scanResult, err := dbClient.Scan(emptyContext, &dynamodb.ScanInput{
		TableName:                 &dynamodbTable,
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return errors.Errorf("Could not scan DynamoDB table: %s", err.Error())
	}
	if scanResult.Count == 0 {
		return errors.Errorf("Could not find DynamoDB item")
	}
	for _, item := range scanResult.Items {
		lockId := item["LockID"]
		log.Printf("Item: %+v", item)
		log.Printf("Deleting item with LockId: %+v", lockId)
		//_, err := dbClient.DeleteItem(emptyContext,
		//	&dynamodb.DeleteItemInput{
		//		TableName: &dynamodbTable,
		//		Key: map[string]dbtypes.AttributeValue{
		//			"LockID": &dbtypes.AttributeValueMemberS{Value: lockId},
		//		},
		//	})
		if err != nil {
			return err
		}

	}

	return nil
}

func extractVariables(workspace string) (string, string, error) {
	// TODO Extract quay_user and major version
	if strings.Contains(workspace, ":") {
		// Normalized RUN_ID
		keyParts := strings.Split(workspace, ":")
		return keyParts[0], keyParts[1], nil
	} else {
		// Non-normalized RUN_ID
		for _, scenario := range scenarios {
			if strings.Contains(workspace, scenario.Name) {
				keyIdLength := len(workspace) - len(scenario.Name) - 1
				return workspace[:keyIdLength], scenario.Name, nil
			}
		}
	}
	return "", "", errors.Errorf("Could not extract keyId or scenarioName from s3 object")
}

func getS3Objects(cfg aws.Config) []s3types.Object {
	s3Client := s3.NewFromConfig(cfg)
	s3Output, err := s3Client.ListObjectsV2(emptyContext, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3Bucket),
	})
	if err != nil {
		log.Fatalf("Could not list s3 bucket: %s", err)
	}
	return s3Output.Contents
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
