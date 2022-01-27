package cmd

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const histFileName = ".gocd_history"

var histfilepath = filepath.Join(os.Getenv("HOME"), histFileName)

var (
	cdFazzyFlag   bool
	cdGHQFlag     bool
	cdHistoryFlag bool
)

func existsDir(d string) bool {
	if info, err := os.Stat(d); !os.IsNotExist(err) {
		return info.IsDir()
	}
	return false
}

func saveCurrentDir() error {
	pwd := os.Getenv("PWD")

	f, err := os.OpenFile(histfilepath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := &bytes.Buffer{}
	io.Copy(buf, f)
	f.Seek(0, io.SeekStart)
	bs := bufio.NewScanner(buf)
	skips := map[string]struct{}{}
	var fsize int64
	for bs.Scan() {
		line := bs.Text()
		if line == pwd || line == "\n" || !existsDir(line) {
			continue
		}
		if _, ok := skips[line]; ok {
			continue
		}

		skips[line] = struct{}{}
		n, _ := fmt.Fprintln(f, line)
		fsize += int64(n)
	}
	n, _ := fmt.Fprintln(f, pwd)
	fsize += int64(n)
	os.Truncate(histfilepath, fsize)

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
			target = "."
		} else {
			target = args[0]
		}
		fmt.Printf("\\cd \"$(find %s ! -readable -prune -o -type d | fzf -0 || echo .)\"", target)
	case cdGHQFlag:
		fmt.Printf("GOCD_ROOT=$(ghq root);\n")
		fmt.Printf("GOCD_TARGET=$(ghq list | fzf);\n")
		fmt.Printf("echo $GOCD_TARGET;\n")
		fmt.Printf("[ \"${GOCD_TARGET}\" = \"\" ] || \\cd $GOCD_ROOT/$GOCD_TARGET;\n")
	case cdHistoryFlag:
		fmt.Printf("\\cd \"$(tac ~/.gocd_history | fzf || echo .)\"")
	default:
		var target string
		if len(args) == 0 {
			target = os.Getenv("HOME")
		} else {
			target = args[0]
		}
		fmt.Printf("\\cd \"%s\"", target)
	}
	return nil
}

func init() {
	cdFlag := flag.NewFlagSet("cd", flag.ExitOnError)
	cdFlag.BoolVar(&cdFazzyFlag, "f", false, "")
	cdFlag.BoolVar(&cdGHQFlag, "ghq", false, "")
	cdFlag.BoolVar(&cdHistoryFlag, "h", false, "")
	addSubCommand(command{
		flagSet: cdFlag,
		action:  cdCmd,
	})
}
