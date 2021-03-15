package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type command struct {
	flagSet *flag.FlagSet
	action  func(args []string) error
}

var subcmds = map[string]command{}

func register(cmd command) {
	subcmds[cmd.flagSet.Name()] = cmd
}

func handle(name string, args []string) error {
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

func initCmd(args []string) error {
	fmt.Printf("%s", strings.TrimSpace(`
gocd() {
    eval "$(go-cd cd $@)"
}
_gocd() {
    local cur prev word cword
    _init_completion || return
    compopt -o filenames
    case $cword in
    1)
        COMPREPLY=( $(compgen -W "-f -ghq -h" -- $cur) $(compgen -d -- $cur ) );;
    *)
        COMPREPLY=( $(compgen -d -- $cur) );;
    esac
}
alias cd="gocd"
complete -o nosort -F _gocd gocd
complete -o nosort -F _gocd cd
`))
	return nil
}

var (
	cdFazzyFlag   bool
	cdGHQFlag     bool
	cdHistoryFlag bool
)

func saveCurrentDir() error {
	home := os.Getenv("HOME")
	pwd := os.Getenv("PWD")
	historyDir := filepath.Join(home, ".gocd_history")

	f, err := os.OpenFile(historyDir, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := &bytes.Buffer{}
	io.Copy(buf, f)
	f.Seek(0, io.SeekStart)
	bs := bufio.NewScanner(buf)
	for i := 0; bs.Scan() && i < 30; {
		line := bs.Text()
		if line == pwd || line == "\n" {
			continue
		}
		fmt.Fprintln(f, line)
		i++
	}
	fmt.Fprintln(f, pwd)
	return nil
}

func cdCmd(args []string) error {
	if err := saveCurrentDir(); err != nil {
		return err
	}

	switch {
	case cdFazzyFlag:
		var target string
		if len(args) == 0 {
			target = os.Getenv("PWD")
		} else {
			target = args[0]
		}
		fmt.Printf("\\cd $(find %s -type d | fzf --height 40%% --reverse --preview='')", target)
	case cdGHQFlag:
		fmt.Printf("GOCD_ROOT=$(ghq root);\n")
		fmt.Printf("GOCD_TARGET=$(ghq list | fzf --height 40%% --reverse --preview='');\n")
		fmt.Printf("echo $GOCD_TARGET;\n")
		fmt.Printf("[ \"${GOCD_TARGET}\" = \"\" ] || \\cd $GOCD_ROOT/$GOCD_TARGET;\n")
	case cdHistoryFlag:
		fmt.Printf("\\cd $(tac ~/.gocd_history | fzf --height 40%% --reverse --preview='' || echo .)")
	default:
		var target string
		if len(args) == 0 {
			target = os.Getenv("HOME")
		} else {
			target = args[0]
		}
		fmt.Printf("\\cd %s", target)
	}
	return nil
}

func init() {
	register(command{
		flagSet: flag.NewFlagSet("init", flag.ExitOnError),
		action:  initCmd,
	})

	cdFlag := flag.NewFlagSet("cd", flag.ExitOnError)
	cdFlag.BoolVar(&cdFazzyFlag, "f", false, "")
	cdFlag.BoolVar(&cdGHQFlag, "ghq", false, "")
	cdFlag.BoolVar(&cdHistoryFlag, "h", false, "")
	register(command{
		flagSet: cdFlag,
		action:  cdCmd,
	})
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 0 {
		if err := handle(args[0], args[1:]); err != nil {
			log.Fatal(err)
		}
	}
}
