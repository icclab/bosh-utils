package main

import (
	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
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

func grep(searchTerm string) {
	r, _ := ch.FindByPartialName(searchTerm)
	if verbose {
		for i := 0; i < len(r.Credentials); i++ {
			c, _ := ch.GetLatestValue(r.Credentials[i].Name)
			fmt.Println(r.Credentials[i].Name+":", color.YellowString(string(c.Value)))
		}
	} else {
		for i := 0; i < len(r.Credentials); i++ {
			fmt.Println(r.Credentials[i].Name)
		}
	}
}

func getVar() {
	c, _ := ch.GetLatestValue(os.Args[1])

	//If can't find directly, search all paths
	if c.Name == "" {
		r, _ := ch.FindByPartialName(os.Args[1])
		switch len(r.Credentials) {
		case 0:
			fmt.Println(color.HiRedString(os.Args[1]), "not found under any path")
			return
		case 1:
			c, _ = ch.GetLatestValue(r.Credentials[0].Name)
		default:
			grep(os.Args[1])
			return
		}
	}

	//If we found only one matching value then copy to clipboard and print according to required verbosity
	_ = clipboard.WriteAll(string(c.Value))
	if xVerbose {
		fmt.Println(color.GreenString("ID:"), c.Id)
		fmt.Println(color.GreenString("Base:"), c.Base)
		fmt.Println(color.GreenString("Metadata:"), c.Metadata)
		fmt.Println(color.GreenString("Name:"), color.RedString(c.Name))
		fmt.Println(color.GreenString("Type:"), c.Type)
		fmt.Println(color.GreenString("Value:"), color.YellowString(string(c.Value)))
		fmt.Println(color.GreenString("Creation Date:"), c.VersionCreatedAt)

		fmt.Println("Value copied clipboard!")
	} else if verbose {
		fmt.Println(color.RedString(c.Name)+":", color.YellowString(string(c.Value)), "(Copied!)")
	} else {
		fmt.Println(color.RedString(c.Name), "copied to clipboard!")
	}
}
