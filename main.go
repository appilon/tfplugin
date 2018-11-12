package main

import (
	"log"
	"os"

	"github.com/appilon/tfplugin/cmd/docs"
	"github.com/appilon/tfplugin/cmd/schema"
	"github.com/appilon/tfplugin/cmd/upgrade/golang"
	"github.com/appilon/tfplugin/cmd/upgrade/pr"
	"github.com/appilon/tfplugin/cmd/upgrade/sdk"
	"github.com/mitchellh/cli"
)

func main() {
	log.SetFlags(log.Llongfile)
	c := cli.NewCLI("tfplugin", "0.1.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		schema.CommandName: schema.CommandFactory,
		docs.CommandName:   docs.CommandFactory,
		golang.CommandName: golang.CommandFactory,
		sdk.CommandName:    sdk.CommandFactory,
		pr.CommandName:     pr.CommandFactory,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
