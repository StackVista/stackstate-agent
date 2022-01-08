package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	cfgFile            string
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
	cobra.OnInitialize(initConfig)

	// persistent flags global for the application
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.beest.yaml)")

	rootCmd.PersistentFlags().StringVar(&yardId, YardIdFlag, "", "used as prefix for resource names")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".beest" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".beest")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func commonVariables() map[string]interface{} {
	if yardId == "" {
		yardId = scenario.Name
	}
	return map[string]interface{}{
		"yard_id": fmt.Sprintf("beest-%s-%s", yardId, agentCurrentBranch),
	}
}
