package sdk

import (
	"flag"

	"github.com/mitchellh/cli"
)

const CommandName = "upgrade sdk"

type command struct{}

func (c *command) Help() string {
	return ""
}

func (c *command) Synopsis() string {
	return ""
}

func CommandFactory() (cli.Command, error) {
	return &command{}, nil
}

func (c *command) Run(args []string) int {
	flags := flag.NewFlagSet(CommandName, flag.ExitOnError)
	flags.Parse(args)
	return 0
}
