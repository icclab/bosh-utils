package main

import (
	"bufio"
	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"os"
	"sort"
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

	if os.Args[1] == "backup" {
		backup("")
	} else {

		getVar()
	}
}

func printHelp() {
	fmt.Println("Usage: [VAR_NAME]\nUsage: [VAR_NAME] -v (Shows the value)\nUsage: [VAR_NAME] -V (Shows all the details for this secret)")
	fmt.Println("\nIf a single, exactly or partially matching secret is found, then it is copied to clipboard. If not, then a search is done using the provided name," +
		"and the names of any secrets found are printed.")
	fmt.Println("\nNote that '-V' only works when one match is found, however '-v' will still show the values even if many matches are found.")
}

func grep(creds []credentials.Base) {
	//Sort so that vars of each deployment are grouped together
	sort.Slice(creds, func(i, j int) bool { return creds[i].Name < creds[j].Name })

	//Set colors to use
	deploymentColor := color.New(color.FgGreen)
	varValueColor := color.New(color.FgYellow)
	varNameColor := color.New(color.FgCyan)

	verbose = verbose || xVerbose //Show values if either is set
	currDeployment := ""
	for i := 0; i < len(creds); i++ {

		switch strings.Count(creds[i].Name, "/") {
		//Var unrelated to a bosh or a specific deployment
		case 1:
			if currDeployment != "/" {
				currDeployment = "/"
				_, _ = deploymentColor.Println(currDeployment + ":")
			}

			if verbose {
				c, _ := ch.GetLatestVersion(creds[i].Name)
				varNameColor.Print("  " + c.Name[1:] + ": ")
				_, _ = varValueColor.Println(getVarString(c.Name, c.Type, 4))
			} else {
				_, _ = varNameColor.Println("  " + creds[i].Name[1:])
			}

		//Var part of a deployment
		case 3:
			parts := strings.Split(creds[i].Name, "/")
			if parts[2] != currDeployment {
				currDeployment = parts[2]
				_, _ = deploymentColor.Println(currDeployment + ":")
			}

			if verbose {
				c, _ := ch.GetLatestVersion(creds[i].Name)
				varNameColor.Print("  " + parts[3] + ": ")
				_, _ = varValueColor.Println(getVarString(c.Name, c.Type, 4))
			} else {
				_, _ = varNameColor.Println("  " + parts[3])
			}
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
			fmt.Println(os.Args[1], "not found under any path")
			return
		case 1:
			c, _ = ch.GetLatestVersion(r.Credentials[0].Name)
		default:
			grep(r.Credentials)
			return
		}
	}

	//If we found only one matching value then copy to clipboard and print according to required verbosity
	varString := getVarString(c.Name, c.Type, 4)
	copied := false
	if c.Type == "value" || c.Type == "password" {
		_ = clipboard.WriteAll(varString)
		copied = true
	}

	if xVerbose {
		fmt.Println(color.GreenString("ID:"), c.Id)
		fmt.Println(color.GreenString("Base:"), c.Base)
		fmt.Println(color.GreenString("Metadata:"), c.Metadata)
		fmt.Println(color.GreenString("Name:"), color.CyanString(c.Name))
		fmt.Println(color.GreenString("Type:"), c.Type)
		fmt.Println(color.GreenString("Value:"), color.YellowString(varString))
		fmt.Println(color.GreenString("Creation Date:"), c.VersionCreatedAt)
	} else if verbose || !copied {
		fmt.Println(color.CyanString(c.Name)+":", color.YellowString(varString))
	} else {
		fmt.Println(color.CyanString(c.Name), "copied to clipboard!")
		return
	}

	if copied {
		fmt.Println("Value copied clipboard!")
	}
}

func backup(dep string) {
	results, _ := ch.FindByPartialName("")
	creds := results.Credentials

	f, _ := os.Create("bosh-backup.yml")
	defer f.Close()
	writer := bufio.NewWriter(f)

	writer.WriteString("credentials:\n")
	for i := 0; i < len(creds); i++ {
		cred, _ := ch.GetLatestVersion(creds[i].Name)

		writer.WriteString("- name: " + cred.Name + "\n")
		writer.WriteString("  type: " + cred.Type + "\n")
		val := getVarString(cred.Name, cred.Type, 4)

		if cred.Type == "certificate" || cred.Type == "rsa" || cred.Type == "ssh" {
			writer.WriteString("  value:" + val + "\n")
		} else if cred.Type == "password" || cred.Type == "value" {
			writer.WriteString("  value: " + val + "\n")
		} else if cred.Type == "user" {
			writer.WriteString("  value: " + strings.Replace(val, "\n", "\n  ", -1) + "\n")
		}
	}
	writer.Flush()
}

func getVarString(name string, varType string, baseIndent int) string {
	spaces := strings.Repeat(" ", baseIndent)

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
		return fmt.Sprintf("\n"+spaces+"ca: |\n"+spaces+"  %s\n"+spaces+"certificate: |\n"+spaces+"  %s\n"+spaces+"private_key: |\n"+spaces+"  %s",
			strings.Replace(cert.Ca, "\n", "\n"+spaces+"  ", -1),
			strings.Replace(cert.Certificate, "\n", "\n"+spaces+"  ", -1),
			strings.Replace(cert.PrivateKey, "\n", "\n"+spaces+"  ", -1))
	case "json":
		v, _ := ch.GetLatestJSON(name)
		return fmt.Sprintf("\n%v", v.Value)
	case "rsa":
		v, _ := ch.GetLatestRSA(name)
		rsa := v.Value
		return fmt.Sprintf("\n"+spaces+"public_key: |\n"+spaces+"  %s\n"+spaces+"private_key: |\n"+spaces+"  %s",
			strings.Replace(rsa.PublicKey, "\n", "\n"+spaces+"  ", -1),
			strings.Replace(rsa.PrivateKey, "\n", "\n"+spaces+"  ", -1))
	case "ssh":
		v, _ := ch.GetLatestSSH(name)
		ssh := v.Value
		return fmt.Sprintf("\n"+spaces+"public_key: |\n"+spaces+"  %s\n"+spaces+"private_key: |\n"+spaces+"  %s",
			strings.Replace(ssh.PublicKey, "\n", "\n"+spaces+"  ", -1),
			strings.Replace(ssh.PrivateKey, "\n", "\n"+spaces+"  ", -1))
	case "user":
		v, _ := ch.GetLatestUser(name)
		user := v.Value
		return fmt.Sprintf("\n  username: %s\n  password: %s", user.Username, user.Password)
	default:
		return ""
	}
}
