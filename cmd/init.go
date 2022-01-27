package cmd

import (
	"flag"
	"fmt"
	"strings"
)

func initCmd(args []string) error {
	fmt.Printf("%s", strings.TrimSpace(`
gocd() {
    source <(go-cd cd "$@")
}
alias cd="gocd"
`))
	return nil
}

func init() {
	addSubCommand(command{
		flagSet: flag.NewFlagSet("init", flag.ExitOnError),
		action:  initCmd,
	})
}
