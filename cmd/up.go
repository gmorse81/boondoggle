package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/viper"

	"gopkg.in/yaml.v2"

	"github.com/gmorse81/mwg-env-switcher/boondoggle/pkg"
	"github.com/spf13/cobra"
)

var namespace string
var release string

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Do all the things.",
	Long: `boondoggle up with no extra flags will configure your defaults and deploy using helm. 
	Flags can be used to change configuration based on your needs.`,
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

		// Add the imagePullSecrets using kubectl
		err = b.AddImagePullSecret()
		if err != nil {
			return err
		}

		// Clone any projects that need to be cloned.
		err = b.DoClone()
		if err != nil {
			return err
		}

		// Build the containers that need to be built.
		err = b.DoBuild()
		if err != nil {
			return err
		}

		// Run helm dep up
		err = b.DepUp()
		if err != nil {
			return err
		}

		// Run the helm upgrade --install command
		err = b.DoUpgrade()
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	upCmd.Flags().StringVar(&release, "release", "", "The helm release name")
	upCmd.MarkFlagRequired("release")
	viper.BindPFlag("release", upCmd.Flags().Lookup("release"))

	upCmd.Flags().StringVar(&namespace, "namespace", "", "The kubernetes namespace of this release")
	viper.BindPFlag("namespace", upCmd.Flags().Lookup("namespace"))

	rootCmd.AddCommand(upCmd)
}
