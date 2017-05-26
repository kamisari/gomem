package main

import (
	"bufio"
	"fmt"
	"github.com/kamisari/gomem"
	"io"
	"os"
	"strings"
)

///bare test

// for Read and confirm
// accept exchange output and input
// default: writer = os.Stdout
//        : reader = os.Stdin
var (
	igs         *gomem.Gomems
	interWriter io.Writer = os.Stdout
	interReader io.Reader = os.Stdin
)

// simple read
func read(msg string) string {
	fmt.Print(msg)
	sc := bufio.NewScanner(interReader)
	sc.Scan()
	if sc.Err() != nil {
		panic(sc.Err())
	}
	return sc.Text()
}

// simple confirm
func confirm(msg string) bool {
	fmt.Fprint(interWriter, msg+" [yes:no]?>")
	for sc, i := bufio.NewScanner(interReader), 0; sc.Scan() && i < 2; i++ {
		if sc.Err() != nil {
			panic(sc.Err)
		}
		switch sc.Text() {
		case "yes", "y":
			return true
		case "no", "n":
			return false
		default:
			fmt.Fprintln(interWriter, sc.Text())
			fmt.Fprint(interWriter, msg+" [yes:no]?>")
		}
	}
	return false
}

/// commands
func la() (string, error) {
	var str string
	for key, v := range igs.Gmap {
		str += fmt.Sprintln("-----", key, "-----")
		str += fmt.Sprintln("[", v.JSON.Title, "]")
		str += fmt.Sprintln(v.JSON.Content)
	}
	return str, nil
}
func ls() (string, error) {
	var str string
	for key := range igs.Gmap {
		str += fmt.Sprintln(key)
	}
	return str, nil
}
func newGomem() (string, error) {
	fpath := read("filename:>")
	g, err := gomem.New(fpath, true)
	if err != nil {
		return err.Error(), nil
	}
	g.JSON.Title = read("title:>")
	g.JSON.Content = read("content:>")
	if err := igs.AddGomem(g); err != nil {
		return err.Error(), nil
	}
	return "new gomem key:" + fpath, nil
}
func newGomemWithName(s string) (string, error) {
	g, err := gomem.New(s, true)
	if err != nil {
		return err.Error(), nil
	}
	g.JSON.Title = read("title:>")
	g.JSON.Content = read("content:>")
	if err := igs.AddGomem(g); err != nil {
		return err.Error(), nil
	}
	return "new gomem included", nil
}
func writeAll() (string, error) {
	b := confirm(fmt.Sprintf("write mems into %s", igs.GetDir()))
	var result string
	if b {
		for key, x := range igs.Gmap {
			if err := x.WriteFile(); err != nil {
				result += fmt.Sprintln(key, err.Error())
			}
		}
		return result, nil
	}
	return "", nil
}
func state() (string, error) {
	var str string
	str += fmt.Sprintln("igs.dir:", igs.GetDir())
	for key, v := range igs.Gmap {
		str += fmt.Sprintln("----------", key, "----------")
		str += fmt.Sprintln("override:", v.Override)
	}
	return str, nil
}
func cd() (string, error) {
	pwd := igs.GetDir()
	if err := os.Chdir(read("cd path:>")); err != nil {
		return err.Error(), nil
	}
	tmpgs, err := gomem.GomemsNew()
	if err != nil {
		if err := os.Chdir(pwd); err != nil {
			return err.Error(), nil
		}
		return err.Error(), nil
	}
	igs = tmpgs
	return "changed directory to:" + igs.GetDir(), nil
}
func show(s string) (string, error) {
	if !strings.HasSuffix(s, ".json") {
		s = s + ".json"
	}
	g, ok := igs.Gmap[s]
	if !ok {
		return "not found:" + s, nil
	}
	str := fmt.Sprintln("[", g.JSON.Title, "]")
	str += fmt.Sprintln(g.JSON.Content)
	return str, nil
}
func remove(s string) (string, error) {
	if !strings.HasSuffix(s, ".json") {
		s = s + ".json"
	}
	fullpath, err := igs.GetAbs(s)
	if err != nil {
		return err.Error(), nil
	}
	err = os.Remove(fullpath)
	if err != nil {
		return err.Error(), nil
	}
	delete(igs.Gmap, s)
	return fullpath + " is removed", nil
}

// Interactive make interactive session
// this file export only this function
func interactive(r io.Reader, w io.Writer, prefix string, gs *gomem.Gomems) error {
	if gs == nil || gs.Gmap == nil {
		return fmt.Errorf("gs or gs.Gmap is nil, exit session")
	}
	igs = gs
	interReader = r
	interWriter = w

	sub := gomem.SubNewWithBase(r, w)
	sub.Addf("la", la, "show Gmap")
	sub.Addf("ls", ls, "ls Gmap keys")
	sub.Addf("new", newGomem, "new gomem")
	sub.Addf("write", writeAll, "write all data to gs.dir")
	sub.Addf("state", state, "show state of gs")
	sub.Addf("cd", cd, "change working directory")
	sub.Addf(":q", gomem.Exit, "exit alias")

	sub.Addfa("show", show, "show title and content")
	sub.Addfa("remove", remove, "remove json cache")
	sub.Addfa("new", newGomemWithName, "")

	if err := sub.Repl(prefix); err != nil {
		return err
	}
	return nil
}

// mock TODO: impl
// specify not interactive then run the this function
func run(w io.Writer, gs *gomem.Gomems) error {
	fmt.Fprintln(w, "debug: run")
	return nil
}
