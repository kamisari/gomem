package gomem

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type subcmd struct {
	f       func() (string, error)
	fa      func(string) (string, error)
	helpmsg string
}

// SubCommands interp functions for Repl
type SubCommands struct {
	w   io.Writer
	r   io.Reader
	Map map[string]*subcmd
}

// ErrValidExit for valid exit, for Repl
var ErrValidExit = errors.New("valid exit")

// Repl is Read Eval Print Loop
// call function in SubCommands[string]
// string is from os.Stdin
// if return ErrValidExit then return nil
func (sub *SubCommands) Repl(prefix string) error {
	sc := bufio.NewScanner(sub.r)
	for {
		fmt.Fprint(sub.w, prefix)
		if !sc.Scan() {
			return fmt.Errorf("fail sc.Scan")
		}
		if sc.Err() != nil {
			return sc.Err()
		}
		cmdline := strings.SplitN(strings.TrimSpace(sc.Text()), " ", 2)
		cmd, ok := sub.Map[strings.TrimSpace(cmdline[0])]
		if !ok {
			fmt.Fprintf(sub.w, "invalid subcommand: %q\n", sc.Text())
			continue
		}
		var result string
		var err error
		switch {
		case cmd.f != nil:
			result, err = cmd.f()
		case cmd.fa != nil && len(cmdline) == 2:
			result, err = cmd.fa(strings.TrimSpace(cmdline[1]))
		default:
			fmt.Fprintf(sub.w, "invalid subcommand: argument: %q\n", cmdline)
			continue
		}
		if err != nil {
			switch err {
			case ErrValidExit:
				fmt.Fprintln(sub.w, result) // exit message
				return nil
			default:
				return err
			}
		}
		fmt.Fprintln(sub.w, result)
	}
}

// Addf append function
func (sub *SubCommands) Addf(key string, fnc func() (string, error), help string) {
	sub.Map[key] = &subcmd{
		f:       fnc,
		helpmsg: help,
	}
}

// Addfa append function with accept argument
func (sub *SubCommands) Addfa(key string, fnc func(string) (string, error), help string) {
	sub.Map[key] = &subcmd{
		fa:      fnc,
		helpmsg: help,
	}
}

// SubNewWithBase return SubCommands with base commands
// for "exit" and "help"
func SubNewWithBase(r io.Reader, w io.Writer) *SubCommands {
	sub := &SubCommands{
		Map: make(map[string]*subcmd),
		r:   r,
		w:   w,
	}
	sub.Addf("exit", Exit, "call exit")
	sub.Addf("help", sub.Help, "show subcommands")
	return sub
}

/// base commmands

// Exit Base Commands for valid exit
func Exit() (string, error) {
	return "", ErrValidExit
}

// Help Base Commands for show help message
func (sub *SubCommands) Help() (string, error) {
	str := fmt.Sprintln("list commands:")
	for key, v := range sub.Map {
		str += fmt.Sprintf("\t%s\n", key)
		str += fmt.Sprintf("\t\t%s\n", v.helpmsg)
	}
	return str, nil
}
