package main

import (
	"flag"
	"log"

	"github.com/entooone/go-cd/cmd"
)

func run() error {
	flag.Parse()
	args := flag.Args()
	if len(args) != 0 {
		if err := cmd.Run(args[0], args[1:]); err != nil {
			return err
		}
	} else {
		cmd.PrintCommands()
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
