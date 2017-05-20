package gomem

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// TODO: maybe deprecated interface{}
type subcmd struct {
	f       func(interface{}) (string, error)
	value   interface{}
	helpmsg string
}

// SubCommands interp functions for Repl
type SubCommands map[string]subcmd

// ErrValidExit for valid exit, for Repl
var ErrValidExit = errors.New("valid exit")

// Repl is Read Eval Print Loop
// call function in SubCommands[string]
// string is from os.Stdin
// if return ErrValidExit then return nil
func (sub SubCommands) Repl(r io.Reader, w io.Writer, prefix string) error {
	sc := bufio.NewScanner(r)
	for {
		fmt.Fprint(w, prefix)
		if !sc.Scan() {
			return fmt.Errorf("fail sc.Scan")
		}
		if sc.Err() != nil {
			return sc.Err()
		}
		cmd, ok := sub[strings.TrimSpace(sc.Text())]
		if !ok {
			fmt.Fprintf(w, "invalid subcommand: %q\n", sc.Text())
			continue
		}
		result, err := cmd.f(cmd.value)
		if err != nil {
			switch err {
			case ErrValidExit:
				fmt.Fprint(w, result) // exit message
				return nil
			default:
				return err
			}
		}
		fmt.Fprint(w, result)
	}
}

// Addf append function
func (sub SubCommands) Addf(key string, v interface{}, fnc func(interface{}) (string, error), help string) {
	sub[key] = subcmd{
		f:       fnc,
		value:   v,
		helpmsg: help,
	}
}

// SubNewWithBase return SubCommands with base commands
// for "exit" and "help"
func SubNewWithBase() SubCommands {
	sub := make(SubCommands)
	sub.Addf("exit", nil, Exit, "call exit")
	sub.Addf("help", nil, sub.Help, "show subcommands")
	return sub
}

/* base commmands */
// RECONSIDER: move to extended files

// Exit Base Commands for valid exit
func Exit(v interface{}) (string, error) {
	return "", ErrValidExit
}

// Help Base Commands for show help message
func (sub SubCommands) Help(v interface{}) (string, error) {
	str := fmt.Sprintln("list commands:")
	for key, v := range sub {
		str += fmt.Sprintf("\t%s\n", key)
		str += fmt.Sprintf("\t\t%s\n", v.helpmsg)
	}
	return str, nil
}
