package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/viper"

	"gopkg.in/yaml.v2"

	"github.com/gmorse81/boondoggle/boondoggle"
	"github.com/spf13/cobra"
)

var namespace string
var release string

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Runs 'helm upgrade --install' with config based on flags and the contents of boondoggle.yml",
	Long: `boondoggle up with no extra flags will configure your defaults and deploy using helm.
	Flags can be used to change configuration based on your needs.`,
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

		// Add any helm repos that are not already added.
		err = b.AddHelmRepos()
		if err != nil {
			return err
		}

		// Add the imagePullSecrets using kubectl
		err = b.AddImagePullSecret(viper.GetString("namespace"))
		if err != nil {
			return err
		}

		// Clone any projects that need to be cloned.
		err = b.DoClone()
		if err != nil {
			return err
		}

		// Build the containers that need to be built.
		if skipDocker != true {
			err = b.DoPreDeploySteps()
			if err != nil {
				return err
			}
			err = b.DoBuild()
			if err != nil {
				return err
			}
		}

		// Run helm dep up
		err = b.DepUp()
		if err != nil {
			return err
		}

		// Run the helm upgrade --install command
		out, err = b.DoUpgrade(viper.GetString("namespace"), viper.GetString("release"), viper.GetBool("dry-run"), viper.GetBool("helm-secrets"))
		if err != nil {
			return fmt.Errorf("Helm upgrade command reported error: %s", string(out))
		}
		fmt.Println(string(out))

		if skipDocker != true {
			err = b.DoPostDeploySteps()
			if err != nil {
				return err
			}
			err = b.DoPostDeployExec(viper.GetString("namespace"))
			if err != nil {
				return err
			}
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
