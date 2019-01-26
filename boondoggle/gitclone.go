package boondoggle

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"

	"gopkg.in/src-d/go-git.v4"
)

// DoClone determines which projects are being worked locally and clones them if necessary.
func (b *Boondoggle) DoClone() error {
	for _, service := range b.Services {
		// Only do these steps if the repo is running locally and a gitrepo is specified.
		if service.Repository == "localdev" && service.Gitrepo != "" {
			_, err := os.Stat(service.Path)
			if os.IsNotExist(err) {
				s := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
				sshKey, err := ioutil.ReadFile(s)
				if err != nil {
					return err
				}
				client, err := ssh.NewPublicKeys("git", sshKey, "")
				if err != nil {
					return err
				}
				_, err = git.PlainClone(service.Path, false, &git.CloneOptions{
					URL:      service.Gitrepo,
					Progress: os.Stdout,
					Auth:     client,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
