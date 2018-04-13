package boondoggle

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// RawBoondoggle is the struct representation of the boondoggle.yml config file.
type RawBoondoggle struct {
	PullSecretsName string `mapstructure:"pull-secrets-name"`
	HelmRepos       []struct {
		Name            string `mapstructure:"name"`
		URL             string `mapstructure:"url"`
		Promptbasicauth bool   `mapstructure:"promptbasicauth,omitempty"`
		Username        string `mapstructure:"username,omitempty"`
		Password        string `mapstructure:"password,omitempty"`
	} `mapstructure:"helm-repos"`
	Umbrella struct {
		Name         string `mapstructure:"name"`
		Repository   string `mapstructure:"repository"`
		Path         string `mapstructure:"path"`
		Environments []struct {
			Name   string   `mapstructure:"name"`
			Files  []string `mapstructure:"files,omitempty"`
			Values []string `mapstructure:"values,omitempty"`
		} `mapstructure:"environments"`
	} `mapstructure:"umbrella"`
	Services []struct {
		Name               string `mapstructure:"name"`
		Path               string `mapstructure:"path"`
		Gitrepo            string `mapstructure:"gitrepo"`
		Alias              string `mapstructure:"alias,omitempty"`
		Chart              string `mapstructure:"chart"`
		DepValuesAllStates struct {
			Condition    string        `mapstructure:"condition,omitempty"`
			Tags         []string      `mapstructure:"tags,omitempty"`
			Enabled      bool          `mapstructure:"enabled,omitempty"`
			Importvalues []interface{} `mapstructure:"importvalues,omitempty"`
		} `mapstructure:"dep-values-all-states,omitempty"`
		States []struct {
			StateName      string        `mapstructure:"state-name"`
			ContainerBuild string        `mapstructure:"container-build,omitempty"`
			Repository     string        `mapstructure:"repository"`
			HelmValues     []string      `mapstructure:"helm-values,omitempty"`
			Version        string        `mapstructure:"version"`
			Condition      string        `mapstructure:"condition,omitempty"`
			Tags           []string      `mapstructure:"tags,omitempty"`
			Enabled        bool          `mapstructure:"enabled,omitempty"`
			Importvalues   []interface{} `mapstructure:"importvalues,omitempty"`
		} `mapstructure:"states"`
	} `mapstructure:"services"`
}

// Boondoggle is the processed version of RawBoondoggle. It represents the settings of only the chosen options.
type Boondoggle struct {
	PullSecretsName string
	HelmRepos       []HelmRepo
	Umbrella        Umbrella
	Services        []Service
}

// HelmRepo is the data needed to add a Helm Repository. Part of Boondoggle struct.
type HelmRepo struct {
	Name            string
	URL             string
	Promptbasicauth bool
	Username        string
	Password        string
}

// Umbrella is the definition of a Helm Umbrella chart. Part of Boondoggle struct.
type Umbrella struct {
	Name       string
	Path       string
	Repository string
	Files      []string
	Values     []string
}

// Service is the definition of a service (an umbrella dependency). Part of Boondoggle struct.
type Service struct {
	Name           string
	Path           string
	Gitrepo        string
	Alias          string
	Chart          string
	ContainerBuild string
	Repository     string
	HelmValues     []string
	Version        string
	Condition      string
	Tags           []string
	Enabled        bool
	Importvalues   []interface{}
}

// NewBoondoggle unmarshals the boondoggle.yml to RawBoondoggle and returns a processed Boondoggle struct type.
func NewBoondoggle() Boondoggle {
	var config RawBoondoggle
	var boondoggle Boondoggle
	viper.Unmarshal(&config)
	boondoggle.configureServices(config)
	boondoggle.configureUmbrella(config)
	boondoggle.configureTopLevel(config)
	return boondoggle
}

// GetHelmDepName is a helper to return the Alias if there is one, else returns Chart
func (s Service) GetHelmDepName() string {
	if s.Alias != "" {
		return s.Alias
	}
	return s.Chart
}

func (b *Boondoggle) configureTopLevel(r RawBoondoggle) {
	b.PullSecretsName = r.PullSecretsName
	for _, helmrepo := range r.HelmRepos {
		var repoDetails = HelmRepo{
			Name:            helmrepo.Name,
			URL:             helmrepo.URL,
			Promptbasicauth: helmrepo.Promptbasicauth,
			Username:        escapableEnvVarReplace(helmrepo.Username),
			Password:        escapableEnvVarReplace(helmrepo.Password),
		}
		b.HelmRepos = append(b.HelmRepos, repoDetails)
	}
}

// converts RawBoondoggle into the umbrella configurations for Boondoggle
func (b *Boondoggle) configureUmbrella(r RawBoondoggle) {
	umbrellaEnv := viper.GetString("environment")
	var umbrellaEnvKey int
	var err error
	// get the environments slice that matches the requested environment, or default if nothing is provided.
	if umbrellaEnv != "" {
		umbrellaEnvKey, err = getRawUmbrellaEnvkeyByName(umbrellaEnv, r)
	} else {
		umbrellaEnvKey, err = getRawUmbrellaEnvkeyByName("default", r)
	}

	if err != nil {
		// indicates there was not a match for the given environment
		fmt.Print(err)
	} else {
		// build the environment in Boondoggle
		b.Umbrella.Name = r.Umbrella.Name
		b.Umbrella.Path = r.Umbrella.Path
		b.Umbrella.Repository = r.Umbrella.Repository
		b.Umbrella.Values = escapableEnvVarReplaceSlice(r.Umbrella.Environments[umbrellaEnvKey].Values)
		b.Umbrella.Files = r.Umbrella.Environments[umbrellaEnvKey].Files
	}
}

func getRawUmbrellaEnvkeyByName(desiredEnvName string, r RawBoondoggle) (int, error) {
	for key, env := range r.Umbrella.Environments {
		if env.Name == desiredEnvName {
			return key, nil
		}
	}
	return 999, fmt.Errorf("The specified environment was not found")
}

// Converts a RawBoondoggle into the services for Boondoggle so they can be consumed by the rest of the application.
func (b *Boondoggle) configureServices(r RawBoondoggle) {
	// First get the service-state overrides provided by the user in a way that we can work with.
	serviceStates := getServiceStatesMap()
	// For each of the services on RawBoondoggle...
	for _, rawService := range r.Services {
		var chosenStateKey int
		var err error
		if serviceStates[rawService.Name] != "" {
			// if we have a override in the serviceStates, get the state values based on that name.
			chosenStateKey, err = getRawStateKeyByName(rawService.Name, serviceStates[rawService.Name], r)
		} else {
			// if not, find the default
			chosenStateKey, err = getRawStateKeyByName(rawService.Name, "default", r)
		}

		if err != nil {
			// indicates there was not a match for the given service and state-name
			fmt.Print(err)
		} else {
			// build the service from the selected state
			var completeService = Service{
				Name:           rawService.Name,
				Path:           rawService.Path,
				Gitrepo:        rawService.Gitrepo,
				Alias:          rawService.Alias,
				Chart:          rawService.Chart,
				ContainerBuild: escapableEnvVarReplace(rawService.States[chosenStateKey].ContainerBuild),
				Repository:     rawService.States[chosenStateKey].Repository,
				HelmValues:     escapableEnvVarReplaceSlice(rawService.States[chosenStateKey].HelmValues),
				Version:        rawService.States[chosenStateKey].Version,
				Condition:      rawService.States[chosenStateKey].Condition,
				Tags:           rawService.States[chosenStateKey].Tags,
				Enabled:        rawService.States[chosenStateKey].Enabled,
				Importvalues:   rawService.States[chosenStateKey].Importvalues,
			}
			// add the dep-values-all-states if the originals are empty
			if completeService.Condition == "" {
				completeService.Condition = rawService.DepValuesAllStates.Condition
			}
			if len(completeService.Tags) < 1 {
				completeService.Tags = rawService.DepValuesAllStates.Tags
			}
			if completeService.Enabled == false {
				completeService.Enabled = rawService.DepValuesAllStates.Enabled
			}
			if completeService.Importvalues == nil {
				completeService.Importvalues = rawService.DepValuesAllStates.Importvalues
			}

			// Append to Boondoggle.Services
			b.Services = append(b.Services, completeService)
		}
	}
}

// returns a mapping of the --service-state flags from the user's command.
func getServiceStatesMap() map[string]string {
	var serviceStatesMap = make(map[string]string)
	for _, value := range viper.GetStringSlice("service-state") {
		// --service-state flag is formatted "name=value", split on the "=" and return a map of the values.
		splitServiceState := make([]string, 2)
		splitServiceState = strings.Split(value, "=")
		serviceStatesMap[splitServiceState[0]] = splitServiceState[1]
	}
	return serviceStatesMap
}

// returns the key of the chosen state when given a service and a state-name.
func getRawStateKeyByName(desiredServiceName string, desiredServiceState string, b RawBoondoggle) (int, error) {
	// Find the "state_name" matching the desiredServiceState for desiredServiceName
	for _, rawService := range b.Services {
		// if this is the service we want to be working with...
		if rawService.Name == desiredServiceName {
			// loop through the states
			for key, rawState := range rawService.States {
				// if this is the state we want to be working with
				if rawState.StateName == desiredServiceState {
					return key, nil
				}
			}
		}
	}
	return 999, fmt.Errorf("A service or state requested was not found for %s %s", desiredServiceName, desiredServiceState)
}

//replace env vars in a string slice.
func escapableEnvVarReplaceSlice(s []string) []string {
	for key, val := range s {
		s[key] = escapableEnvVarReplace(val)
	}
	return s
}

//escapableEnvVarReplace wraps os.Getenv to allow for escaping with $$.
func escapableEnvVarReplace(s string) string {
	return os.Expand(s, func(s string) string {
		if s == "$" {
			return "$"
		}
		return os.Getenv(s)
	})
}
