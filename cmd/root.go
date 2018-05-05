package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "boondoggle",
	Short: "Boondoggle is a helm umbrella chart preprocessor, a dependency state management tool as well as a local development tool.",
}

var serviceState []string
var umbrellaEnv string
var stateValueOverride []string
var dryRun bool
var setStateAll string
var skipDocker bool

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (./boondoggle.yml)")

	rootCmd.PersistentFlags().StringSliceVarP(&serviceState, "service-state", "s", []string{""}, "Sets a services state eg. my-service=local. Defaults to the 'default' state.")
	viper.BindPFlag("service-state", rootCmd.PersistentFlags().Lookup("service-state"))

	rootCmd.PersistentFlags().StringVarP(&umbrellaEnv, "environment", "e", "default", "Selects the umbrella environment. Defaults to the environment with name: default in the boondoggle.yml file.")
	viper.BindPFlag("environment", rootCmd.PersistentFlags().Lookup("environment"))

	rootCmd.PersistentFlags().StringSliceVarP(&stateValueOverride, "state-v-override", "o", []string{""}, "Override a services's version for the state specified. eg. my-service=1.0.0")
	viper.BindPFlag("state-v-override", rootCmd.PersistentFlags().Lookup("state-v-override"))

	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run will do all steps except for the helm deploy. The helm command that would have been run will be printed.")
	viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))

	rootCmd.PersistentFlags().StringVarP(&setStateAll, "set-state-all", "a", "", "Sets all services to the same state.")
	viper.BindPFlag("set-state-all", rootCmd.PersistentFlags().Lookup("set-state-all"))

	rootCmd.PersistentFlags().BoolVarP(&skipDocker, "skip-docker", "k", false, "Skips the docker build step.")
	viper.BindPFlag("skip-docker", rootCmd.PersistentFlags().Lookup("skip-docker"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in pwd with name "boondoggle.yml" (without extension).
		viper.SetConfigName("boondoggle")
		viper.AddConfigPath(".")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
	}
}
