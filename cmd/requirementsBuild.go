package cmd

import (
	"log"
	"os"

	"github.com/gmorse81/boondoggle/v3/boondoggle"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fastReq bool

// requirementsBuildCmd represents the requirementsBuild command
var requirementsBuildCmd = &cobra.Command{
	Use:   "requirements-build",
	Short: "Only build the requirements.yaml file and run helm dep up",
	Long: `This command will build the dependencies in requirements.yaml and then run helm dep up.
No deployment or container builds will occur.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		// Get a NewBoondoggle built from config.
		var config boondoggle.RawBoondoggle
		viper.Unmarshal(&config)
		b := boondoggle.NewBoondoggle(config, viper.GetString("environment"), viper.GetString("set-state-all"), viper.GetStringSlice("service-state"), map[string]string{}, log.New(os.Stdout, "", 0))

		//Build requirements.yml
		r := boondoggle.BuildRequirements(b, viper.GetStringSlice("state-v-override"))

		// Write the new requirements.yml
		err := boondoggle.WriteRequirements(r, b)
		if err != nil {
			return err
		}

		if !fastReq {
			// Add any helm repos that are not already added.
			err = b.AddHelmRepos()
			if err != nil {
				return err
			}

			// Run helm dep up
			err = b.DepUp()
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	requirementsBuildCmd.Flags().BoolVar(&fastReq, "fast", false, "Build the requirements file, but do not download the dependencies")
	viper.BindPFlag("fast", requirementsBuildCmd.Flags().Lookup("fast"))

	rootCmd.AddCommand(requirementsBuildCmd)
}
