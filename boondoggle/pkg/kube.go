package boondoggle

import (
	"fmt"
	"os/exec"
	"strings"

	sshterminal "golang.org/x/crypto/ssh/terminal"
)

//AddImagePullSecret ensures the kubernetes imagePullSecret is set with kubectl.
func (b *Boondoggle) AddImagePullSecret() error {
	if b.PullSecretsName != "" { // if boondoggle config specifies a pullsecretsname
		// Determine if it's already set up.
		in := fmt.Sprintf("get secrets %s", b.PullSecretsName)
		inslice := strings.Split(in, " ")
		cmd := exec.Command("kubectl", inslice...)
		out, err := cmd.CombinedOutput()
		if err != nil && strings.Contains(string(out), "NotFound") {
			// if there was an error and the output of the command contains "NotFound", setup the credentials
			fmt.Println("You need to set up docker hub integration with kubernetes.")
			fmt.Println("Please enter your docker hub EMAIL ADDRESS:")
			var email string
			_, err := fmt.Scanln(&email)
			if err != nil {
				return fmt.Errorf("error with kubectl create secret: %s", err)
			}
			fmt.Println("Please enter your docker hub USERNAME:")
			var username string
			_, err = fmt.Scanln(&username)
			if err != nil {
				return fmt.Errorf("error with kubectl create secret: %s", err)
			}
			fmt.Println("Please enter your docker hub PASSWORD:")
			password, err := sshterminal.ReadPassword(0)
			if err != nil {
				return fmt.Errorf("error with kubectl create secret: %s", err)
			}
			in := fmt.Sprintf("create secret docker-registry %s --docker-username=%s --docker-password=%s --docker-email=%s", b.PullSecretsName, username, password, email)
			inslice := strings.Split(in, " ")
			cmd := exec.Command("kubectl", inslice...)
			out, err := cmd.CombinedOutput()
			fmt.Println(string(out))
			if err != nil {
				return fmt.Errorf("error with kubectl create secret: %s", err)
			}
		} else if err != nil { // else if some other error has happened, return the error.
			return fmt.Errorf("error with kubectl get: %s", err)
		}
		// Otherwise it already exists and we do nothing.
	}
	return nil
}
