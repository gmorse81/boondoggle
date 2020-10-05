package boondoggle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RawBoondoggle is the struct representation of the boondoggle.yml config file.
type RawBoondoggle struct {
	HelmVersion     int    `mapstructure:"helmVersion,omitempty"`
	PullSecretsName string `mapstructure:"pull-secrets-name,omitempty"`
	DockerUsername  string `mapstructure:"docker_username,omitempty"`
	DockerPassword  string `mapstructure:"docker_password,omitempty"`
	DockerEmail     string `mapstructure:"docker_email,omitempty"`
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
			PreDeploySteps []struct {
				Cmd  string   `mapstructure:"cmd,omitempty"`
				Args []string `mapstructure:"args,omitempty"`
			} `mapstructure:"preDeploySteps,omitempty"`
			PostDeploySteps []struct {
				Cmd  string   `mapstructure:"cmd,omitempty"`
				Args []string `mapstructure:"args,omitempty"`
			} `mapstructure:"postDeploySteps,omitempty"`
			PostDeployExec []struct {
				App       string   `mapstructure:"app,omitempty"`
				Container string   `mapstructure:"container,omitempty"`
				Args      []string `mapstructure:"args,omitempty"`
			} `mapstructure:"postDeployExec,omitempty"`
		} `mapstructure:"states"`
	} `mapstructure:"services"`
}

// Boondoggle is the processed version of RawBoondoggle. It represents the settings of only the chosen options.
type Boondoggle struct {
	PullSecretsName string
	DockerUsername  string
	DockerPassword  string
	DockerEmail     string
	HelmVersion     int
	HelmRepos       []HelmRepo
	Umbrella        Umbrella
	Services        []Service
	ExtraEnv        map[string]string
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

// Step contains instructions for a pre, post or post exec build step for local.
type Step struct {
	App       string
	Container string
	Cmd       string
	Args      []string
}

// Service is the definition of a service (an umbrella dependency). Part of Boondoggle struct.
type Service struct {
	Name            string
	Path            string
	Gitrepo         string
	Alias           string
	Chart           string
	ContainerBuild  string
	Repository      string
	HelmValues      []string
	Version         string
	Condition       string
	Tags            []string
	Enabled         bool
	Importvalues    []interface{}
	PreDeploySteps  []Step
	PostDeploySteps []Step
	PostDeployExec  []Step
}

// NewBoondoggle unmarshals the boondoggle.yml to RawBoondoggle and returns a processed Boondoggle struct type.
func NewBoondoggle(config RawBoondoggle, environment string, setStateAll string, serviceState []string, extraEnv map[string]string) Boondoggle {
	var boondoggle Boondoggle
	boondoggle.ExtraEnv = extraEnv
	boondoggle.configureServices(config, setStateAll, serviceState)
	boondoggle.configureUmbrella(config, environment)
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
	b.DockerEmail = b.escapableEnvVarReplace(r.DockerEmail)
	b.DockerPassword = b.escapableEnvVarReplace(r.DockerPassword)
	b.DockerUsername = b.escapableEnvVarReplace(r.DockerUsername)
	if r.HelmVersion == 0 {
		b.HelmVersion = 2
	} else {
		b.HelmVersion = r.HelmVersion
	}
	for _, helmrepo := range r.HelmRepos {
		var repoDetails = HelmRepo{
			Name:            helmrepo.Name,
			URL:             helmrepo.URL,
			Promptbasicauth: helmrepo.Promptbasicauth,
			Username:        b.escapableEnvVarReplace(helmrepo.Username),
			Password:        b.escapableEnvVarReplace(helmrepo.Password),
		}
		b.HelmRepos = append(b.HelmRepos, repoDetails)
	}
}

// converts RawBoondoggle into the umbrella configurations for Boondoggle
func (b *Boondoggle) configureUmbrella(r RawBoondoggle, environment string) {
	umbrellaEnv := environment
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
		b.Umbrella.Path, _ = filepath.Abs(r.Umbrella.Path)
		b.Umbrella.Repository = r.Umbrella.Repository
		b.Umbrella.Values = b.escapableEnvVarReplaceSlice(r.Umbrella.Environments[umbrellaEnvKey].Values)
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
func (b *Boondoggle) configureServices(r RawBoondoggle, setStateAll string, serviceState []string) {
	// First get the service-state overrides provided by the user in a way that we can work with.
	serviceStates := getServiceStatesMap(serviceState)
	// For each of the services on RawBoondoggle...
	for _, rawService := range r.Services {
		var chosenStateKey int
		var err error

		// If the set-state-all flag was not set, set service states with default logic.
		overrideServiceState := setStateAll
		if overrideServiceState == "" {
			if serviceStates[rawService.Name] != "" {
				// if we have a override in the serviceStates, get the state values based on that name.
				chosenStateKey, err = getRawStateKeyByName(rawService.Name, serviceStates[rawService.Name], r)
			} else {
				// if not, find the default
				chosenStateKey, err = getRawStateKeyByName(rawService.Name, "default", r)
			}
		} else {
			chosenStateKey, err = getRawStateKeyByName(rawService.Name, overrideServiceState, r)
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
				ContainerBuild: b.escapableEnvVarReplace(rawService.States[chosenStateKey].ContainerBuild),
				Repository:     rawService.States[chosenStateKey].Repository,
				HelmValues:     b.escapableEnvVarReplaceSlice(rawService.States[chosenStateKey].HelmValues),
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

			// Add the pre and post steps
			if len(rawService.States[chosenStateKey].PreDeploySteps) > 0 {
				for _, val := range rawService.States[chosenStateKey].PreDeploySteps {
					completeService.PreDeploySteps = append(completeService.PreDeploySteps, Step{
						Cmd:  val.Cmd,
						Args: b.escapableEnvVarReplaceSlice(val.Args),
					})
				}
			}

			if len(rawService.States[chosenStateKey].PostDeploySteps) > 0 {
				for _, val := range rawService.States[chosenStateKey].PostDeploySteps {
					completeService.PostDeploySteps = append(completeService.PostDeploySteps, Step{
						Cmd:  val.Cmd,
						Args: b.escapableEnvVarReplaceSlice(val.Args),
					})
				}
			}

			if len(rawService.States[chosenStateKey].PostDeployExec) > 0 {
				for _, val := range rawService.States[chosenStateKey].PostDeployExec {
					completeService.PostDeployExec = append(completeService.PostDeployExec, Step{
						App:       val.App,
						Container: val.Container,
						Args:      b.escapableEnvVarReplaceSlice(val.Args),
					})
				}
			}

			// Append to Boondoggle.Services
			b.Services = append(b.Services, completeService)
		}
	}
}

// returns a mapping of the --service-state flags from the user's command.
func getServiceStatesMap(serviceState []string) map[string]string {
	var serviceStatesMap = make(map[string]string)
	for _, value := range serviceState {
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
func (b *Boondoggle) escapableEnvVarReplaceSlice(s []string) []string {
	for key, val := range s {
		s[key] = b.escapableEnvVarReplace(val)
	}
	return s
}

//escapableEnvVarReplace wraps os.Getenv to allow for escaping with $$.
//populates from either the system's environment variables or Boondoggle.ExtraEnv
func (b *Boondoggle) escapableEnvVarReplace(s string) string {
	return os.Expand(s, func(s string) string {
		if s == "$" {
			return "$"
		}

		realEnvVal := os.Getenv(s)
		extraEnvVal := ""
		for key, val := range b.ExtraEnv {
			if s == key {
				extraEnvVal = val
			}
		}

		if extraEnvVal != "" {
			return extraEnvVal
		}

		return realEnvVal
	})
}
