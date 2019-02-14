package boondoggle

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	sshterminal "golang.org/x/crypto/ssh/terminal"
)

//This file contains the helm commands run by boondoggle using values from Boondoggle

// DoUpgrade builds and runs the helm upgrade --install command.
func (b *Boondoggle) DoUpgrade(namespace string, release string, dryRun bool, useSecrets bool, tls bool, tillerNamespace string) ([]byte, error) {
	fullcommand := []string{"upgrade", "-i"}

	// Add the release name
	if release != "" {
		fullcommand = append(fullcommand, release)
	}

	// Add the umbrella path
	fullcommand = append(fullcommand, b.Umbrella.Path)

	// Add files from the umbrella declartion
	for _, file := range b.Umbrella.Files {
		chunk := fmt.Sprintf("-f %s/%s", b.Umbrella.Path, file)
		fullcommand = append(fullcommand, strings.Split(chunk, " ")...)
	}

	//Set global.projectLocation to the location of the boondoggle.yaml file.
	//This can be used to map volumes for local dev.
	projectLocation := fmt.Sprintf("--set global.projectLocation=%s", os.Getenv("PWD"))
	fullcommand = append(fullcommand, strings.Split(projectLocation, " ")...)

	// Add values from the umbrella declaration
	for _, value := range b.Umbrella.Values {
		chunk := fmt.Sprintf("--set %s", value)
		fullcommand = append(fullcommand, strings.Split(chunk, " ")...)
	}

	// Add values from each service, append the service's chart name(or alias if supplied)
	for _, service := range b.Services {
		for _, servicevalue := range service.HelmValues {
			chunk := fmt.Sprintf("--set %s.%s", service.GetHelmDepName(), servicevalue)
			fullcommand = append(fullcommand, strings.Split(chunk, " ")...)
		}
	}

	// For services running in local dev, add the cachebuster
	for _, service := range b.Services {
		if service.Repository == "localdev" {
			now := time.Now()
			chunk := fmt.Sprintf("--set %s.boondoggleCacheBust='%d'", service.GetHelmDepName(), now.Unix())
			fullcommand = append(fullcommand, strings.Split(chunk, " ")...)
		}
	}

	// Add the namespace if there is one.
	if namespace != "" {
		chunk := fmt.Sprintf("--namespace %s", namespace)
		fullcommand = append(fullcommand, strings.Split(chunk, " ")...)
	}

	// Add a longer timeout
	chunk := "--timeout 1800 --wait"
	fullcommand = append(fullcommand, strings.Split(chunk, " ")...)

	// Add Tiller namespace
	if tillerNamespace != "kube-system" {
		chunk = fmt.Sprintf("--tiller-namespace %s", tillerNamespace)
		fullcommand = append(fullcommand, strings.Split(chunk, " ")...)
	}

	// Add tls flag
	if tls {
		fullcommand = append(fullcommand, "--tls")
	}

	if useSecrets {
		fullcommand = append([]string{"secrets"}, fullcommand...)
	}

	cmd := exec.Command("helm", fullcommand...)

	// Run the command
	if dryRun == false {
		fmt.Println("Installing the environment...")
		out, err := cmd.CombinedOutput()
		return out, err
	}

	return []byte(fmt.Sprintf("%s", cmd.Args)), nil

}

//DepUp runs "helm dependency update".
func (b *Boondoggle) DepUp() error {
	cmd := exec.Command("helm", "dep", "up", b.Umbrella.Path)
	_, err := cmd.CombinedOutput()
	fmt.Println("Updating dependencies...")
	if err != nil {
		return fmt.Errorf("There was an error updating the dependencies on the umbrella: %s", err)
	}
	return nil
}

/*
AddHelmRepos uses `helm repo add` to setup the repos listed in boondoggle config.
If promtbasicauth is true, it will prompt the user for the helm repo username and password.
If "username" and "password" are provided, it will use these as the basic auth username and password. environment var replacement is supported for these.
If the result for the environment variable lookup is empty, it will fall back to prompting for username and password.
It will not do anything if the repo is already added.
*/
func (b *Boondoggle) AddHelmRepos() error {
	cmd := exec.Command("helm", "repo", "list")
	out, err := cmd.CombinedOutput()
	fmt.Println("Adding helm repos...")
	if err != nil {
		return fmt.Errorf("error in boondoggle fetching the existing helm chart repos: %s", err)
	}
	for _, repo := range b.HelmRepos {
		// Not the best implementation, but helm does not have a json output for helm repo list.
		// If the output of "helm repo list" does not contain the repo name(by basic string search), add it.
		if strings.Contains(string(out), repo.Name) == false {
			if repo.Promptbasicauth == true { //if the repo requires username and password, prompt for that.

				// get the username either from user input or from boondoggle.yml
				var username string
				if repo.Username == "" {
					fmt.Printf("Enter the username for repo %s: \n", repo.Name)
					_, err := fmt.Scanln(&username)
					if err != nil {
						return fmt.Errorf("error with collecting username for helm chart repo: %s", err)
					}
				} else {
					username = repo.Username
				}

				// get the password either from user input or from boondoggle.yml
				var password string
				if repo.Password == "" {
					fmt.Printf("Enter the password for %s: \n", username)
					bytePassword, err := sshterminal.ReadPassword(0)
					if err != nil {
						return fmt.Errorf("error with collecting password for helm chart repo: %s", err)
					}
					password = string(bytePassword)
				} else {
					password = repo.Password
				}

				u, err := url.Parse(repo.URL)
				if err != nil {
					return fmt.Errorf("error parsing the url of the chart repo when attempting to add to helm: %s", err)
				}

				//Add the basic auth username and password to the URL.
				u.User = url.UserPassword(username, password)
				repoadd(repo.Name, u)
			} else { // else, add without the prompt for username and password.
				u, err := url.Parse(repo.URL)
				if err != nil {
					return fmt.Errorf("error parsing the url of the chart repo: %s", err)
				}
				err = repoadd(repo.Name, u)
				if err != nil {
					return fmt.Errorf("error when trying to add a chart repo to helm registry: %s", err)
				}
			}
		}
	}
	return nil
}

func repoadd(name string, u *url.URL) error {
	cmd := exec.Command("helm", "repo", "add", name, u.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error adding a repo to the helm repository: %s", string(out))
	}
	return nil
}

// SelfFetch will fetch the umbrella chart listed in the boondoggle.yml file.
func (b *Boondoggle) SelfFetch(path string, version string) error {
	cleanRepo := strings.TrimPrefix(b.Umbrella.Repository, "@")
	var fetchcommand string
	if version == "" {
		fetchcommand = fmt.Sprintf("fetch %s/%s --untar -d %s", cleanRepo, b.Umbrella.Name, path)
	} else {
		fetchcommand = fmt.Sprintf("fetch %s/%s --untar --version=%s -d %s", cleanRepo, b.Umbrella.Name, version, path)
	}
	cmd := exec.Command("helm", strings.Split(fetchcommand, " ")...)
	out, err := cmd.CombinedOutput()
	fmt.Println("Fetching the umbrella...")
	if err != nil {
		return fmt.Errorf("error with self fetch: %s", string(out))
	}
	return nil
}
