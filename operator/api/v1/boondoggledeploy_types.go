/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BoondoggleDeploySpec defines the desired state of BoondoggleDeploy
type BoondoggleDeploySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Optional full boondoggle config file. This will be pulled from the umbrella repo's mainline branch if not specified
	// +optional
	RawBoondoggle         RawBoondoggle `json:"rawBoondoggle,omitempty"`
	BoondoggleEnvironment string
	StateVersionOverrides []string
	UseHelmSecrets        bool
	EnvVars               map[string]string
}

type BoondoggleLastProgression string

const (
	ProcessingConfig     BoondoggleLastProgression = "Processing configuration"
	GenerateRequirements BoondoggleLastProgression = "Generating Requirements for Umbrella"
	AddHelmRepos         BoondoggleLastProgression = "Adding Helm Repositories"
	DownloadDependencies BoondoggleLastProgression = "Downloading Dependencies"
	Installing           BoondoggleLastProgression = "Installing"
)

// BoondoggleDeployStatus defines the observed state of BoondoggleDeploy
type BoondoggleDeployStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DeployStatus         release.Status
	BoondoggleLastStatus BoondoggleLastProgression
	LastDeployTimestamp  metav1.Time
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BoondoggleDeploy is the Schema for the boondoggledeploys API
type BoondoggleDeploy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BoondoggleDeploySpec   `json:"spec,omitempty"`
	Status BoondoggleDeployStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BoondoggleDeployList contains a list of BoondoggleDeploy
type BoondoggleDeployList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BoondoggleDeploy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BoondoggleDeploy{}, &BoondoggleDeployList{})
}

type RawBoondoggle struct {
	HelmVersion     int         `json:"helmVersion,omitempty"`
	PullSecretsName string      `json:"pull-secrets-name,omitempty"`
	DockerUsername  string      `json:"docker_username,omitempty"`
	DockerPassword  string      `json:"docker_password,omitempty"`
	DockerEmail     string      `json:"docker_email,omitempty"`
	HelmRepos       []HelmRepos `json:"helm-repos"`
	Umbrella        Umbrella    `json:"umbrella"`
	Services        []Services  `json:"services"`
}

type HelmRepos struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	Promptbasicauth bool   `json:"promptbasicauth,omitempty"`
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
}

type Umbrella struct {
	Name         string         `json:"name"`
	Repository   string         `json:"repository"`
	Path         string         `json:"path"`
	Environments []Environments `json:"environments"`
}

type Environments struct {
	Name           string   `json:"name"`
	Files          []string `json:"files,omitempty"`
	Values         []string `json:"values,omitempty"`
	AddtlHelmFlags []string `json:"addtlHelmFlags,omitempty"`
}

type Services struct {
	Name               string             `json:"name"`
	Path               string             `json:"path"`
	Gitrepo            string             `json:"gitrepo"`
	Alias              string             `json:"alias,omitempty"`
	Chart              string             `json:"chart"`
	DepValuesAllStates DepValuesAllStates `json:"dep-values-all-states,omitempty"`
	States             []States           `json:"states"`
}

type DepValuesAllStates struct {
	Condition string   `json:"condition,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Enabled   bool     `json:"enabled,omitempty"`
}

type States struct {
	StateName      string   `json:"state-name"`
	ContainerBuild string   `json:"container-build,omitempty"`
	Repository     string   `json:"repository"`
	HelmValues     []string `json:"helm-values,omitempty"`
	Version        string   `json:"version"`
	Condition      string   `json:"condition,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Enabled        bool     `json:"enabled,omitempty"`
}
