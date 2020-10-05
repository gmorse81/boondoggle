package boondoggle

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Maintainer describes a Chart maintainer.
type Maintainer struct {
	// Name is a user name or organization name
	Name string `yaml:"name,omitempty"`
	// Email is an optional email address to contact the named maintainer
	Email string `yaml:"email,omitempty"`
	// URL is an optional URL to an address for the named maintainer
	URL string `yaml:"url,omitempty"`
}

// Requirements represents the yaml file requirements.yaml as used in a Helm chart.
type Requirements struct {
	Dependencies []Dependency `yaml:"dependencies"`
}

// Dependency is part of the requirements file
type Dependency struct {
	Name         string        `yaml:"name,omitempty"`
	Version      string        `yaml:"version,omitempty"`
	Repository   string        `yaml:"repository,omitempty"`
	Condition    string        `yaml:"condition,omitempty"`
	Tags         []string      `yaml:"tags,omitempty"`
	Enabled      bool          `yaml:"enabled,omitempty"`
	Importvalues []interface{} `yaml:"importvalues,omitempty"`
	Alias        string        `yaml:"alias,omitempty"`
}

// Chart for a Chart file. This models the structure of a Chart.yaml file.
type Chart struct {
	Name         string            `yaml:"name,omitempty"`
	Home         string            `yaml:"home,omitempty"`
	Sources      []string          `yaml:"sources,omitempty"`
	Version      string            `yaml:"version,omitempty"`
	Description  string            `yaml:"description,omitempty"`
	Keywords     []string          `yaml:"keywords,omitempty"`
	Maintainers  []Maintainer      `yaml:"maintainers,omitempty"`
	Icon         string            `yaml:"icon,omitempty"`
	APIVersion   string            `yaml:"apiVersion,omitempty"`
	Condition    string            `yaml:"condition,omitempty"`
	Tags         string            `yaml:"tags,omitempty"`
	AppVersion   string            `yaml:"appVersion,omitempty"`
	Deprecated   bool              `yaml:"deprecated,omitempty"`
	Annotations  map[string]string `yaml:"annotations,omitempty"`
	KubeVersion  string            `yaml:"kubeVersion,omitempty"`
	Dependencies []Dependency      `yaml:"dependencies,omitempty"`
	Type         string            `yaml:"type,omitempty"`
}

//BuildRequirements converts a Boondoggle into a Helm Requirements struct.
func BuildRequirements(b Boondoggle, svo []string) Requirements {
	var r Requirements
	var repoLocation string
	for _, service := range b.Services {

		if service.Repository == "localdev" {
			repoLocation = fmt.Sprintf("file://%s/%s/%s", os.Getenv("PWD"), service.Path, service.Chart)
		} else {
			repoLocation = service.Repository
		}

		version := getVersionFlag(service, svo)

		var dependency = Dependency{
			Name:         service.Chart,
			Version:      version,
			Repository:   repoLocation,
			Condition:    service.Condition,
			Tags:         service.Tags,
			Enabled:      service.Enabled,
			Importvalues: service.Importvalues,
			Alias:        service.Alias,
		}

		r.Dependencies = append(r.Dependencies, dependency)
	}
	return r
}

// WriteRequirements writes the dependencies to eithe chart.yaml or requirements.yaml depending on the version of helm you are running
func WriteRequirements(r Requirements, b Boondoggle) error {
	// read the chart.yaml file
	chart := Chart{}
	chartBytes, err := ioutil.ReadFile(b.Umbrella.Path + "/Chart.yaml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(chartBytes, &chart)
	if err != nil {
		return err
	}
	// determine if chart is helm2 or helm3
	if chart.APIVersion == "v2" {
		chart.Dependencies = r.Dependencies
		out, err := yaml.Marshal(chart)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("%s/Chart.yaml", b.Umbrella.Path)
		err = ioutil.WriteFile(path, out, 0644)
		if err != nil {
			return err
		}
		return nil
	}

	if chart.APIVersion == "v1" {
		out, err := yaml.Marshal(r)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("%s/requirements.yaml", b.Umbrella.Path)
		err = ioutil.WriteFile(path, out, 0644)
		if err != nil {
			return err
		}
		return nil
	}

	// write requirements to correct location based on version
	return fmt.Errorf("invalid chart APIVersion specified")
}

func getVersionFlag(service Service, svo []string) string {
	// for each of the --state-v-override flags...
	for _, override := range svo {
		//Split into slice on the "="
		splitvalue := strings.Split(override, "=")
		// if the left side of the value equals the service name, return the version.
		if splitvalue[0] == service.Name && len(splitvalue) == 2 {
			return splitvalue[1]
		}
	}
	// if no match found, return the original version.
	return service.Version
}

func getLocalRepoLocation(s Service, l string) string {
	var repo string
	repo = fmt.Sprintf("file://%s/%s/%s", l, s.Name, s.Chart)
	return repo
}
