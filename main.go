package main

import (
	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"os"
	"strings"
)

var ch *credhub.CredHub
var verbose bool
var xVerbose bool

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
		return
	}

	var err error
	ch, err = credhub.New(strings.Replace(os.Getenv("BOSH_ENVIRONMENT"), "25555", "8844", 1), credhub.CaCerts(os.Getenv("CREDHUB_CA_CERT")),
		credhub.Auth(auth.UaaClientCredentials(os.Getenv("CREDHUB_CLIENT"), os.Getenv("CREDHUB_SECRET"))))

	if err != nil {
		fmt.Println("Failed to set a new Credhub target.", err)
		return
	}

	for _, arg := range os.Args {
		switch arg {
		case "-v":
			verbose = true
		case "-V":
			xVerbose = true
		}
	}

	getVar()
}

func printHelp() {
	fmt.Println("Usage: [VAR_NAME]\nUsage: [VAR_NAME] -v (Shows the value)\nUsage: [VAR_NAME] -V (Shows all the details for this secret)")
	fmt.Println("\nIf a single, exactly matching secret is found, then it is copied to clipboard. If not, then a search is done using the provided name," +
		"and the names of any secrets found are printed.")
	fmt.Println("\nNote that '-V' only works when one match is found, however '-v' will still show the values even if many matches are found.")
}

func grep(searchResults credentials.FindResults) {
	if verbose {
		for i := 0; i < len(searchResults.Credentials); i++ {
			c, _ := ch.GetLatestVersion(searchResults.Credentials[i].Name)
			fmt.Println(color.RedString(c.Name)+":", color.YellowString(getVarString(c.Name, c.Type)))
		}
	} else {
		for i := 0; i < len(searchResults.Credentials); i++ {
			fmt.Println(searchResults.Credentials[i].Name)
		}
	}
}

func getVar() {
	c, _ := ch.GetLatestVersion(os.Args[1])
	//If can't find directly search all paths
	if c.Name == "" {
		r, _ := ch.FindByPartialName(os.Args[1])
		switch len(r.Credentials) {
		case 0:
			fmt.Println(color.RedString(os.Args[1]), "not found under any path")
			return
		case 1:
			c, _ = ch.GetLatestVersion(r.Credentials[0].Name)
		default:
			grep(r)
			return
		}
	}

	//If we found only one matching value then copy to clipboard and print according to required verbosity
	varString := getVarString(c.Name, c.Type)
	copied := false
	if c.Type == "value" || c.Type == "password" {
		_ = clipboard.WriteAll(varString)
		copied = true
	}

	if xVerbose {
		fmt.Println(color.GreenString("ID:"), c.Id)
		fmt.Println(color.GreenString("Base:"), c.Base)
		fmt.Println(color.GreenString("Metadata:"), c.Metadata)
		fmt.Println(color.GreenString("Name:"), color.RedString(c.Name))
		fmt.Println(color.GreenString("Type:"), c.Type)
		fmt.Println(color.GreenString("Value:"), color.YellowString(varString))
		fmt.Println(color.GreenString("Creation Date:"), c.VersionCreatedAt)
	} else if verbose || !copied {
		fmt.Println(color.RedString(c.Name)+":", color.YellowString(varString))
	} else {
		fmt.Println(color.RedString(c.Name), "copied to clipboard!")
		return
	}

	if copied {
		fmt.Println("Value copied clipboard!")
	}
}

func getVarString(name string, varType string) string {
	switch varType {
	case "value":
		v, _ := ch.GetLatestValue(name)
		return string(v.Value)
	case "password":
		v, _ := ch.GetLatestPassword(name)
		return string(v.Value)
	case "certificate":
		v, _ := ch.GetLatestCertificate(name)
		cert := v.Value
		return fmt.Sprintf("\nca: |\n%s\ncertificate: |\n%s\nprivate_key: |\n%s", cert.Ca, cert.Certificate, cert.PrivateKey)
	case "json":
		v, _ := ch.GetLatestJSON(name)
		return fmt.Sprintf("\n%v", v.Value)
	case "rsa":
		v, _ := ch.GetLatestRSA(name)
		rsa := v.Value
		return fmt.Sprintf("\npublic_key: |\n%s\nprivate_key: |\n%s", rsa.PublicKey, rsa.PrivateKey)
	case "ssh":
		v, _ := ch.GetLatestSSH(name)
		ssh := v.Value
		return fmt.Sprintf("\npublic_key: |\n%s\nprivate_key: |\n%s", ssh.PublicKey, ssh.PrivateKey)
	case "user":
		v, _ := ch.GetLatestUser(name)
		user := v.Value
		return fmt.Sprintf("\nusername: %s\npassword: %s", user.Username, user.Password)
	default:
		return ""
	}
}
