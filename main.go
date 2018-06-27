package main

import (
	"log"
	"os"

	"github.com/appilon/tfdev/docs"
	"github.com/appilon/tfdev/schema"
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
