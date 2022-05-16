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
		fullcommand = append(fullcommand, "--set-string", value)
	}

	// Add values from each service, append the service's chart name(or alias if supplied)
	for _, service := range b.Services {
		for _, servicevalue := range service.HelmValues {
			chunk := fmt.Sprintf("%s.%s", service.GetHelmDepName(), servicevalue)
			fullcommand = append(fullcommand, "--set-string", chunk)
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
	if b.is2() {
		chunk := "--timeout 1800 --wait"
		fullcommand = append(fullcommand, strings.Split(chunk, " ")...)
	} else {
		chunk := "--timeout 1800s --wait"
		fullcommand = append(fullcommand, strings.Split(chunk, " ")...)
	}

	// Add additional helm flags
	fullcommand = append(fullcommand, b.Umbrella.AddtlHelmFlags...)

	// Add Tiller namespace
	if b.is2() {
		if tillerNamespace != "kube-system" {
			chunk := fmt.Sprintf("--tiller-namespace %s", tillerNamespace)
			fullcommand = append(fullcommand, strings.Split(chunk, " ")...)
		}
	}

	// Add tls flag
	if b.is2() {
		if tls {
			fullcommand = append(fullcommand, "--tls")
		}
	}

	if b.SuperSecret {
		fullcommand = append(fullcommand, "--debug")
	}

	if useSecrets {
		fullcommand = append([]string{"secrets"}, fullcommand...)
	}

	cmd := exec.Command("helm", fullcommand...)

	// Run the command
	if !dryRun {
		b.L.Print("Installing the environment...")
		if b.Verbose {
			b.L.Print(Format(Cyan, "Command: "+cmd.String()))
		}
		out, err := cmd.CombinedOutput()
		return out, err
	}

	return []byte(fmt.Sprintf("%s", cmd.Args)), nil

}

//DepUp runs "helm dependency update".
func (b *Boondoggle) DepUp() error {
	b.L.Print("Updating dependencies...")
	cmd := exec.Command("helm", "dep", "up", b.Umbrella.Path)
	if b.Verbose {
		b.L.Print(Format(Cyan, "Command: "+cmd.String()))
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("there was an error updating the dependencies on the umbrella: %s", err)
	}
	if b.Verbose {
		b.L.Print(string(out))
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
	if b.Verbose {
		b.L.Print(Format(Cyan, "Command: "+cmd.String()))
	}
	out, _ := cmd.CombinedOutput()
	if b.Verbose {
		b.L.Print(string(out))
	}
	b.L.Print("Adding helm repos...")
	for _, repo := range b.HelmRepos {
		// Not the best implementation, but helm does not have a json output for helm repo list.
		// If the output of "helm repo list" does not contain the repo name(by basic string search), add it.
		if !strings.Contains(string(out), repo.Name) {
			if repo.Promptbasicauth { //if the repo requires username and password, prompt for that.

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
				repoadd(repo.Name, u, b.Verbose, b.L)
			} else { // else, add without the prompt for username and password.
				u, err := url.Parse(repo.URL)
				if err != nil {
					return fmt.Errorf("error parsing the url of the chart repo: %s", err)
				}
				err = repoadd(repo.Name, u, b.Verbose, b.L)
				if err != nil {
					return fmt.Errorf("error when trying to add a chart repo to helm registry: %s", err)
				}
			}
		}
	}
	return nil
}

func repoadd(name string, u *url.URL, verbose bool, logger LogPrinter) error {
	fullcommand := []string{"repo", "add", name, u.String()}
	if verbose {
		fullcommand = append(fullcommand, "--debug")
	}
	cmd := exec.Command("helm", fullcommand...)
	if verbose {
		logger.Print(Format(Cyan, "Command: "+cmd.String()))
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error adding a repo to the helm repository: %s", string(out))
	}
	if verbose {
		logger.Print(string(out))
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
	if b.Verbose {
		b.L.Print(Format(Cyan, "Command: "+cmd.String()))
	}
	out, err := cmd.CombinedOutput()
	b.L.Print("Fetching the umbrella...")
	if b.Verbose {
		b.L.Print(string(out))
	}
	if err != nil {
		return fmt.Errorf("error with self fetch: %s", string(out))
	}
	return nil
}

func (b Boondoggle) is2() bool {
	return b.HelmVersion == 2
}
