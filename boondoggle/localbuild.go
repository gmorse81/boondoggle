package boondoggle

import (
	"os"
	"os/exec"
	"strings"
)

//DoBuild builds the localdev container based on the command in the boondoggle config file.
func (b *Boondoggle) DoBuild() error {
	for _, service := range b.Services {
		// Only do these steps if the repo is running locally and a container-build is specified.
		if service.Repository == "localdev" && service.ContainerBuild != "" {
			cmdslice := strings.Split(service.ContainerBuild, " ")
			cmd := exec.Command("docker", cmdslice...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}
	return nil
}

// DoPreDeploySteps runs the preDeploySteps outlined in the boondoggle.yml file for the services with the state set to "localdev"
// This is used for building any steps that need to happen before deploying a local environment.
func (b *Boondoggle) DoPreDeploySteps() error {
	for _, service := range b.Services {
		if service.Repository == "localdev" && len(service.PreDeploySteps) > 0 {
			for _, step := range service.PreDeploySteps {
				cmd := exec.Command(step.Cmd, step.Args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			}
		}
	}
	return nil
}

// DoPostDeploySteps runs the postDeploySteps outlined in the boondoggle.yml file for the services with the state set to "localdev"
// This is used for building any steps that need to happen after deploying a local environment.
func (b *Boondoggle) DoPostDeploySteps() error {
	for _, service := range b.Services {
		if service.Repository == "localdev" && len(service.PostDeploySteps) > 0 {
			for _, step := range service.PostDeploySteps {
				cmd := exec.Command(step.Cmd, step.Args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			}
		}
	}
	return nil
}

// DoPostDeployExec runs commands in the app outlined in the boondoggle.yml file for the services with the state set to "localdev"
func (b *Boondoggle) DoPostDeployExec(namespace string) error {
	b.L.Print("running post exec...")
	for _, service := range b.Services {
		if service.Repository == "localdev" && len(service.PostDeployExec) > 0 {
			for _, step := range service.PostDeployExec {
				_, err := b.runExecCommand(namespace, step.App, step.Container, step.Args)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// runExecCommand executes the command against the cluster on the specified namespace and app and container combo.
func (b *Boondoggle) runExecCommand(namespace string, appName string, container string, command []string) (string, error) {

	// Get the pod name based on the namespace and the app= tag
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "-o", "go-template", "--template", "{{(index .items 0).metadata.name}}", "--selector", "app="+appName)
	b.L.Print(cmd.Args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	fragmentSlice := []string{"exec", "-n", namespace, "-c", container, string(out), "--"}
	fragmentSlice = append(fragmentSlice, command...)
	cmd = exec.Command("kubectl", fragmentSlice...)
	b.L.Print(cmd.Args)
	out, err = cmd.CombinedOutput()
	b.L.Print(string(out))
	return string(out), err
}
