package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"strings"
	"time"
)

const (
	ScenariosPath = "./scenarios.yml"

	RunIdFlag          = "run-id"
	AssumeYesFlag      = "assume-yes"
	AssumeYesShortFlag = "y"
)

var (
	runId     string
	assumeYes bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "beest",
	Short: "Black-box testing bees",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
		return initializeConfig(cmd)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	defaultRunId := time.Now().Format("20060102150405")
	rootCmd.PersistentFlags().StringVar(&runId, RunIdFlag, defaultRunId, "identifier of this specific execution")
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	// Bind to environment variables
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	// Safely shortens the length of the run id, because it used in yard resource names
	runId = trimTo(runId, 24)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --run-id to RUN_ID
		if strings.Contains(f.Name, "-") {
			envVar := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, envVar)
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func trimTo(s string, maxLength int) string {
	if len(s) < maxLength {
		return s
	}
	return s[:maxLength]
}
