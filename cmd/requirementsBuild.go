package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/gmorse81/boondoggle/boondoggle/pkg"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// requirementsBuildCmd represents the requirementsBuild command
var requirementsBuildCmd = &cobra.Command{
	Use:   "requirements-build",
	Short: "Only build the requirements.yaml file and run helm dep up",
	Long: `This command will build the dependencies in requirements.yaml and then run helm dep up. 
No deployment or container builds will occur.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		// Get a NewBoondoggle built from config.
		b := boondoggle.NewBoondoggle()

		//Build requirements.yml
		r := boondoggle.BuildRequirements(b)

		// Write the new requirements.yml
		out, err := yaml.Marshal(r)
		path := fmt.Sprintf("%s/requirements.yaml", b.Umbrella.Path)
		ioutil.WriteFile(path, out, 0644)
		if err != nil {
			return err
		}

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

		return nil
	},
}

func init() {
	rootCmd.AddCommand(requirementsBuildCmd)
}
