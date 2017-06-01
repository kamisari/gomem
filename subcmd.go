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
	r           io.Reader
	w           io.Writer
	Map         map[string]*subcmd
	Prefix      string
	InterCh     chan string  // accept another input
	callBackCh  *chan string // callBackCh = &CallBackBuf
	CallBackBuf chan string
}

// ErrValidExit for valid exit, for Repl
var ErrValidExit = errors.New("valid exit")

// Repl is Read Eval Print Loop
// call function in SubCommands[string]
// string is from os.Stdin
// if return ErrValidExit then return nil
func (sub *SubCommands) Repl() error {
	sc := bufio.NewScanner(sub.r)
	done := false
	for {
		var s string
		select {
		case inter := <-sub.InterCh:
			s = strings.TrimSpace(inter)
		case callback := <-*sub.callBackCh:
			s = callback
		default:
			if done {
				return nil
			}
			fmt.Fprint(sub.w, sub.Prefix)
			if !sc.Scan() {
				return fmt.Errorf("fail sc.Scan")
			}
			if sc.Err() != nil {
				return sc.Err()
			}
			s = strings.TrimSpace(sc.Text())
		}

		var result string
		var err error
		cmdline := strings.SplitN(s, " ", 2)
		cmd, ok := sub.Map[strings.TrimSpace(cmdline[0])]
		if !ok {
			fmt.Fprintf(sub.w, "invalid subcommand: %q\n", s)
			continue
		}
		switch {
		case cmd.fa != nil && len(cmdline) == 2:
			result, err = cmd.fa(strings.TrimSpace(cmdline[1]))
		case cmd.f != nil:
			result, err = cmd.f()
		default:
			fmt.Fprintf(sub.w, "invalid subcommand: argument: %q\n", cmdline)
			continue
		}
		if err != nil {
			switch err {
			case ErrValidExit:
				if len(sub.CallBackBuf) != 0 {
					// callback
					sub.callBackCh = &sub.CallBackBuf
					done = true
					continue
				}
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
	if _, ok := sub.Map[key]; ok {
		sub.Map[key].f = fnc
		if sub.Map[key].helpmsg == "" {
			sub.Map[key].helpmsg = help
		}
		return
	}
	sub.Map[key] = &subcmd{
		f:       fnc,
		helpmsg: help,
	}
}

// Addfa append function with accept argument
func (sub *SubCommands) Addfa(key string, fnc func(string) (string, error), help string) {
	if _, ok := sub.Map[key]; ok {
		sub.Map[key].fa = fnc
		if sub.Map[key].helpmsg == "" {
			sub.Map[key].helpmsg = help
		}
		return
	}
	sub.Map[key] = &subcmd{
		fa:      fnc,
		helpmsg: help,
	}
}

// SubNew return SubCommands
func SubNew(r io.Reader, w io.Writer) *SubCommands {
	mock := make(chan string)
	sub := &SubCommands{
		r:           r,
		w:           w,
		Map:         make(map[string]*subcmd),
		InterCh:     make(chan string, 1),
		CallBackBuf: make(chan string, 1),
		callBackCh:  &mock,
	}
	return sub
}

/// base commmands

// Exit Base Commands for valid exit
func (sub *SubCommands) Exit() (string, error) {
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
