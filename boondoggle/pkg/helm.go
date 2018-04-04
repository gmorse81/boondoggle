package boondoggle

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"

	sshterminal "golang.org/x/crypto/ssh/terminal"
)

//This file contains the helm commands run by boondoggle using values from Boondoggle

// TODOs
//run the helm upgrade command
//add extra --set for build cachebuster when projects have localdev
//use the alias for --set when there is one, else use reponame
//add the local-files-path value
//run dep up

// helm upgrade mwg --install ${PWD}/mwg-umbrella-chart -f ${PWD}/mwg-umbrella-chart/local.yml --set global.projectLocation=${PWD}

// DoUpgrade builds and runs the helm upgrade --install command.
func (b *Boondoggle) DoUpgrade() error {
	fullcommand := []string{"upgrade", "-i"}

	//Set global.projectLocation to the location of the boondoggle.yaml file.
	//This can be used to map volumes for local dev.
	projectLocation := fmt.Sprintf("--set global.projectLocation=%s", os.Getenv("PWD"))
	fullcommand = append(fullcommand, strings.Split(projectLocation, " ")...)

	for _, file := range b.Umbrella.Files {
		chunk := fmt.Sprintf("-f %s/%s", b.Umbrella.Path, file)
		fullcommand = append(fullcommand, strings.Split(chunk, " ")...)
	}

	fullcommand = append(fullcommand, fmt.Sprintf("./%s", b.Umbrella.Path))

	fmt.Printf("helm %s", strings.Trim(fmt.Sprint(fullcommand), "[]"))
	return nil
}

//DepUp runs "helm dependency update".
func (b *Boondoggle) DepUp() error {
	cmd := exec.Command("helm", "dep", "up", b.Umbrella.Path)
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		return fmt.Errorf("error with helm dep up: %s", err)
	}
	return nil
}

//AddHelmRepos uses `helm repo add` to setup the repos in boondoggle config.
//if promtbasicauth is true, it will prompt the user for the helm repo username and password
//it will not do anything if the repo is already added.
func (b *Boondoggle) AddHelmRepos() error {
	cmd := exec.Command("helm", "repo", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error with helm add repo list: %s", err)
	}
	for _, repo := range b.HelmRepos {
		// Not the best implementation, but helm does not have a json output for helm repo list.
		// If the output of "helm repo list" does not contain the repo name(by basic string search), add it.
		if strings.Contains(string(out), repo.Name) == false {
			if repo.Promptbasicauth == true { //if the repo requires username and password, prompt for that.
				fmt.Printf("Enter the username for repo %s: \n", repo.Name)
				var username string
				_, err := fmt.Scanln(&username)
				if err != nil {
					return fmt.Errorf("error with helm add repo: %s", err)
				}
				fmt.Printf("Enter the password for %s: \n", username)
				password, err := sshterminal.ReadPassword(0)
				if err != nil {
					return fmt.Errorf("error with helm add repo: %s", err)
				}
				u, err := url.Parse(repo.URL)
				if err != nil {
					return fmt.Errorf("error with helm add repo: %s", err)
				}
				//Add the basic auth username and password to the URL.
				u.User = url.UserPassword(username, string(password))
				repoadd(repo.Name, u)
			} else { // else, add without the prompt for username and password.
				u, err := url.Parse(repo.URL)
				if err != nil {
					return fmt.Errorf("error with helm add repo: %s", err)
				}
				err = repoadd(repo.Name, u)
				if err != nil {
					return fmt.Errorf("error with helm add repo: %s", err)
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
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Print(string(out))
	return nil
}
