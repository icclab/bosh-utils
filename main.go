package main

import (
	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"fmt"
	"github.com/atotto/clipboard"
	"os"
)

var ch *credhub.CredHub

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
		return
	}

	var err error
	ch, err = credhub.New("https://192.168.200.6:8844", credhub.CaCerts(os.Getenv("CREDHUB_CA_CERT")),
		credhub.Auth(auth.UaaClientCredentials(os.Getenv("CREDHUB_CLIENT"), os.Getenv("CREDHUB_SECRET"))))

	if err != nil {
		fmt.Println("Failed to set a new Credhub target.", err)
		return
	}

	if os.Args[1] == "grep" {
		grep()
	} else {
		getVar()
	}
}

func printHelp() {
	fmt.Println("Usage: [VAR_NAME]\nUsage: [VAR_NAME] -v (Shows the value)\nUsage: [VAR_NAME] -V (Shows all the details for this secret)")
	fmt.Println("Usage: grep [SEARCH_TERM]")
}

func grep() {
	r, _ := ch.FindByPartialName(os.Args[2])
	for i := 0; i < len(r.Credentials); i++ {
		c, _ := ch.GetLatestValue(r.Credentials[i].Name)
		fmt.Println(r.Credentials[i].Name+":", c.Value)
	}
}

func getVar() {
	c, _ := ch.GetLatestValue(os.Args[1])
	if c.Name == "" {
		return
	}

	_ = clipboard.WriteAll(string(c.Value))
	if len(os.Args) > 2 && os.Args[2] == "-V" {
		fmt.Print("ID: ", c.Id, "\n", "Base: ", c.Base, "\n", "Metadata: ", c.Metadata, "\n",
			"Name: ", c.Name, "\n", "Type: ", c.Type, "\n", "Value: ", c.Value, "\n", "Creation Date: ", c.VersionCreatedAt, "\n")
	} else if len(os.Args) > 2 && os.Args[2] == "-v" {
		fmt.Println(c.Name+":", c.Value)
	} else {
		fmt.Println(c.Name, "copied to clipboard!")
	}
}
