package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

const (
	ScenariosPath = "./scenarios.yml"

	AgentCurrentBranchEnvVar = "AGENT_CURRENT_BRANCH"

	YardIdFlag         = "yard-id"
	AssumeYesFlag      = "assume-yes"
	AssumeYesShortFlag = "y"
)

var (
	agentCurrentBranch string

	yardId    string
	assumeYes bool

	scenario *Scenario
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "beest",
	Short: "Black-box testing bees",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ensureCurrentBranch()

	cobra.CheckErr(rootCmd.Execute())
}

func ensureCurrentBranch() {
	var ok bool
	agentCurrentBranch, ok = os.LookupEnv(AgentCurrentBranchEnvVar)
	if !ok {
		log.Fatalf("Mandatory environment variable missing: %s", AgentCurrentBranchEnvVar)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&yardId, YardIdFlag, "", "used as prefix for resource names")
}

func commonVariables() map[string]interface{} {
	if yardId == "" {
		yardId = scenario.Name
	}
	trimmedBranch := agentCurrentBranch[:24]
	return map[string]interface{}{
		"yard_id": fmt.Sprintf("beest-%s-%s", yardId, trimmedBranch),
	}
}
