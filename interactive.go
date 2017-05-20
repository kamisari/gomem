package gomem

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// for Read and confirm
// accept exchange output and input
var (
	InterWriter io.Writer = os.Stdout
	InterReader io.Reader = os.Stdin
)

/// for interactive
// simple read
func read(msg string) string {
	fmt.Print(msg)
	sc := bufio.NewScanner(InterReader)
	sc.Scan()
	if sc.Err() != nil {
		panic(sc.Err())
	}
	return sc.Text()
}

// simple confirm
func confirm(msg string) bool {
	fmt.Print(msg + " [yes:no]?>")
	for sc, i := bufio.NewScanner(InterReader), 0; sc.Scan() && i < 2; i++ {
		if sc.Err() != nil {
			panic(sc.Err)
		}
		switch sc.Text() {
		case "yes", "y":
			return true
		case "no", "n":
			return false
		default:
			fmt.Fprintln(InterWriter, sc.Text())
			fmt.Fprint(InterWriter, msg + " [yes:no]?>")
		}
	}
	return false
}

// state
func isValidGS(v interface{}) *Gomems {
	gs, ok := v.(*Gomems)
	if !ok {
		panic("isValidGS: invalid interface v is not Gomems")
	}
	if gs == nil {
		panic("isValidGS: uninitialized value")
	}
	return gs
}

/// commands

func show(v interface{}) (string, error) {
	gs := isValidGS(v)
	var str string
	for key, v := range gs.Gmap {
		str += fmt.Sprintln("-----", key, "-----")
		str += fmt.Sprintf("[%s]\n", v.JSON.Title)
		str += fmt.Sprintf("%s\n", v.JSON.Content)
	}
	return str, nil
}

func ls(v interface{}) (string, error) {
	gs := isValidGS(v)
	var str string
	for key := range gs.Gmap {
		str += fmt.Sprintln(key)
	}
	return str, nil
}

// mock
func Interactive(r io.Reader, w io.Writer, gs *Gomems) error {
	if gs == nil || gs.Gmap == nil {
		return fmt.Errorf("gs or gs.Gmap is nil, exit session")
	}
	fmt.Fprintln(w, "debug: interactive")
	fmt.Fprintf(w, "debug:gs:%v\ndebug:\n", gs)
	/// commands TODO: remove

	sub := SubNewWithBase()
	sub.Addf("show",gs,  show, "show Gmap")
	sub.Addf("ls", gs, ls, "ls Gmap keys")
	// TODO: many addf
	//sub.Addf()
	if err := sub.Repl(r, w, "gomem:>"); err != nil {
		return err
	}
	return nil
}
