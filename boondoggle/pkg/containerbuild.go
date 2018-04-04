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
