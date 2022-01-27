package cmd

import (
	"flag"
	"fmt"
	"os"
)

type command struct {
	flagSet *flag.FlagSet
	action  func(args []string) error
}

var subcmds = map[string]command{}

func addSubCommand(cmd command) {
	subcmds[cmd.flagSet.Name()] = cmd
}

func Run(name string, args []string) error {
	cmd, ok := subcmds[name]
	if !ok {
		return fmt.Errorf("sub command '%s' is not found", name)
	}
	if err := cmd.flagSet.Parse(args); err != nil {
		return err
	}

	cmd.action(cmd.flagSet.Args())

	return nil
}

func PrintCommands() {
	fmt.Fprintf(os.Stderr, "usage: %s COMMAND [OPTIONS]\n", os.Args[0])
	for name := range subcmds {
		fmt.Fprintf(os.Stderr, "  %s\n", name)
	}
}
