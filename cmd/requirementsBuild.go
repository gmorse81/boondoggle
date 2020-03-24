package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/gmorse81/boondoggle/boondoggle"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
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
		b := boondoggle.NewBoondoggle(config, viper.GetString("environment"), viper.GetString("set-state-all"), viper.GetStringSlice("service-state"), map[string]string{})

		//Build requirements.yml
		r := boondoggle.BuildRequirements(b, viper.GetStringSlice("state-v-override"))

		// Write the new requirements.yml
		out, err := yaml.Marshal(r)
		path := fmt.Sprintf("%s/requirements.yaml", b.Umbrella.Path)
		ioutil.WriteFile(path, out, 0644)
		if err != nil {
			return err
		}

		if !fastReq {
			// Add any helm repos that are not already added.
			err = b.AddHelmRepos()
			if err != nil {
				return err
			}

			// Clone any projects that need to be cloned.
			err = b.DoClone()
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
