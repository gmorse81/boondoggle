package boondoggle

import (
	"fmt"
	"os/exec"
	"strings"

	sshterminal "golang.org/x/crypto/ssh/terminal"
)

//AddImagePullSecret ensures the kubernetes imagePullSecret is set with kubectl.
func (b *Boondoggle) AddImagePullSecret(namespace string) error {
	if b.PullSecretsName != "" { // if boondoggle config specifies a pullsecretsname

		err := b.CreateNamespaceIfNotExists(namespace)
		if err != nil {
			b.L.Print(err.Error())
		}

		// Determine if it's already set up.
		in := fmt.Sprintf("get secrets %s", b.PullSecretsName)
		inslice := strings.Split(in, " ")

		// If a namesapce was provided in the "up" command, include the namespace in the check.
		// Add the namespace if there is one.
		if namespace != "" {
			chunk := fmt.Sprintf("--namespace %s", namespace)
			inslice = append(inslice, strings.Split(chunk, " ")...)
		}

		cmd := exec.Command("kubectl", inslice...)
		out, err := cmd.CombinedOutput()

		if err != nil && strings.Contains(string(out), "NotFound") {
			// if there was an error and the output of the command contains "NotFound", setup the credentials

			var email string
			if b.DockerEmail == "" {
				fmt.Println("You need to set up docker hub integration with kubernetes.")
				fmt.Println("Please enter your docker hub EMAIL ADDRESS:")
				_, err := fmt.Scanln(&email)
				if err != nil {
					return fmt.Errorf("error with kubectl create secret: %s", err)
				}
			} else {
				email = b.DockerEmail
			}

			var username string
			if b.DockerUsername == "" {
				fmt.Println("Please enter your docker hub USERNAME:")
				_, err = fmt.Scanln(&username)
				if err != nil {
					return fmt.Errorf("error with kubectl create secret: %s", err)
				}
			} else {
				username = b.DockerUsername
			}

			var password []byte
			if b.DockerPassword == "" {
				fmt.Println("Please enter your docker hub PASSWORD:")
				password, err = sshterminal.ReadPassword(0)
				if err != nil {
					return fmt.Errorf("error with kubectl create secret: %s", err)
				}
			} else {
				password = []byte(b.DockerPassword)
			}

			in := fmt.Sprintf("create secret docker-registry %s --docker-username=%s --docker-password=%s --docker-email=%s", b.PullSecretsName, username, password, email)
			inslice := strings.Split(in, " ")

			// If a namesapce was provided in the "up" command, include the namespace when creating the secret.
			// Add the namespace if there is one.
			if namespace != "" {
				chunk := fmt.Sprintf("--namespace %s", namespace)
				inslice = append(inslice, strings.Split(chunk, " ")...)
			}

			cmd := exec.Command("kubectl", inslice...)
			if b.Verbose {
				b.L.Print(Format(Cyan, "Command: "+cmd.String()))
			}
			out, err := cmd.CombinedOutput()
			if b.Verbose {
				b.L.Print(string(out))
			}
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

// CreateNamespaceIfNotExists creates a kubernetes namespace if it does not already exist in the cluster.
func (b *Boondoggle) CreateNamespaceIfNotExists(namespace string) error {
	// Create the namespace in the cluster if there is one provided
	if namespace != "" {
		// check if namespace exists
		checkNamespace := exec.Command("kubectl", "get", "namespace", namespace)
		if b.Verbose {
			b.L.Print(Format(Cyan, "Command: "+checkNamespace.String()))
		}
		out, err := checkNamespace.CombinedOutput()
		if b.Verbose {
			b.L.Print(string(out))
		}
		if err != nil && strings.Contains(string(out), "not found") {
			// if does not exist, create it
			namespaceCommand := exec.Command("kubectl", "create", "namespace", namespace)
			if b.Verbose {
				b.L.Print(Format(Cyan, "Command: "+namespaceCommand.String()))
			}
			out, err := namespaceCommand.CombinedOutput()
			if err != nil {
				return fmt.Errorf("WARN: non-existent namespace could not be created")
			}
			if b.Verbose {
				b.L.Print(string(out))
			}
			b.L.Print("Namespace " + namespace + " created")
		} else {
			b.L.Print("Namespace " + namespace + " already exists. skipping.")
		}
	}
	return nil
}
