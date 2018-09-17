package main

import (
	"log"
	"os"

	"github.com/appilon/tfplugin/cmd/docs"
	"github.com/appilon/tfplugin/cmd/schema"
	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("tfdev", "0.1.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"schema": schema.CommandFactory,
		"docs":   docs.CommandFactory,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
