package gomem

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// for Read and confirm
// accept exchange output and input
// default: writer = os.Stdout
//        : reader = os.Stdin
var (
	igs         *Gomems
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
func show() (string, error) {
	var str string
	for key, v := range igs.Gmap {
		str += fmt.Sprintln("-----", key, "-----")
		str += fmt.Sprintf("[%s]\n", v.JSON.Title)
		str += fmt.Sprintf("%s\n", v.JSON.Content)
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
	flag := confirm("accept override")
	g, err := New(fpath, flag)
	if err != nil {
		return err.Error(), nil
	}
	g.JSON = gJSON{
		Title:   read("title:>"),
		Content: read("content:>"),
	}
	if err := igs.AddGomem(g); err != nil {
		return err.Error(), nil
	}
	return "new gomem included", nil
}
func writeAll() (string, error) {
	b := confirm(fmt.Sprintf("write mems into %s", igs.dir))
	var result string
	if b {
		for key, x := range igs.Gmap {
			err := x.WriteFile()
			result += fmt.Sprintln(key, err.Error())
		}
		return result, nil
	}
	return "", nil
}
func state() (string, error) {
	var str string
	str += fmt.Sprintln("gs.dir:", igs.dir)
	for key, v := range igs.Gmap {
		str += fmt.Sprintln("----------", key, "----------")
		str += fmt.Sprintln("override:", v.Override)
	}
	return str, nil
}

// Interactive make interactive session
// this file export only this function
func Interactive(r io.Reader, w io.Writer, prefix string, gs *Gomems) error {
	if gs == nil || gs.Gmap == nil {
		return fmt.Errorf("gs or gs.Gmap is nil, exit session")
	}
	igs = gs
	interReader = r
	interWriter = w

	sub := SubNewWithBase(r, w)
	sub.Addf("show", show, "show Gmap")
	sub.Addf("ls", ls, "ls Gmap keys")
	sub.Addf("new", newGomem, "new gomem")
	sub.Addf("writeAll", writeAll, "write all data to gs.dir")
	sub.Addf("state", state, "show state of gs")

	if err := sub.Repl(prefix); err != nil {
		return err
	}
	return nil
}
