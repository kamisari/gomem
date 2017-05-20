package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/dory/gomem"
)

type option struct {
	version     bool
	interactive bool
	workdir     string

	// TODO: impl subcmd for run
	subcmd  string
	subarsg []string
}
var opt option

func (opt *option) init() {
	flag.BoolVar(&opt.version, "version", false, "")
	flag.BoolVar(&opt.version, "v", false, "")

	flag.BoolVar(&opt.interactive, "interactive", false, "")
	flag.BoolVar(&opt.interactive, "i", false, "")

	flag.StringVar(&opt.workdir, "workdir", "", "")
	flag.StringVar(&opt.workdir, "w", "", "")
	flag.Parse()
	if flag.NArg() != 0 {
		// TODO: impl parse subcmd
		log.Fatalf("invalid args: %q", flag.Args())
	}
	var err error
	opt.workdir, err = filepath.Abs(opt.workdir)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	log.SetPrefix("gomem:")
	log.SetFlags(log.Lshortfile)

	opt.init()
}

// TODO: impl: interactive(io.Reader, io.Writer, *gomem.Gomems) error
//           : run(io.Writer, *gomem.Gomems, option) error

// LIST: for interactive, list commands
//     : cd, ls, new, show, include, state

/// for interactive
// simple read
func read(msg string) string {
	fmt.Print(msg)
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	if sc.Err() != nil {
		panic(sc.Err())
	}
	return sc.Text()
}

// simple confirm
func confirm(msg string) bool {
	fmt.Print(msg + " [yes:no]?>")
	for sc, i := bufio.NewScanner(os.Stdin), 0; sc.Scan() && i < 2; i++ {
		if sc.Err() != nil {
			log.Fatal(sc.Err())
		}
		switch sc.Text() {
		case "yes", "y":
			return true
		case "no", "n":
			return false
		default:
			fmt.Println(sc.Text())
			fmt.Print(msg + " [yes:no]?>")
		}
	}
	return false
}

// state
func isValidGS(v interface{}) *gomem.Gomems {
	gs, ok := v.(*gomem.Gomems)
	if !ok {
		panic("isValidGS: invalid interface v is not Gomems")
	}
	return gs
}

// mock
func interactive(r io.Reader, w io.Writer, gs *gomem.Gomems) error {
	if gs == nil || gs.Gmap == nil {
		return fmt.Errorf("gs or gs.Gmap is nil, exit session")
	}
	fmt.Fprintln(w, "debug: interactive")
	fmt.Fprintf(w, "debug:gs:%v\ndebug:opt:%v\n", gs, opt)
	/// commands TODO: remove
	show := func(v interface{}) (string, error) {
		gs := isValidGS(v)
		var str string
		for key, v := range gs.Gmap {
			str += fmt.Sprintln("-----", key, "-----")
			str += fmt.Sprintf("[%s]\n", v.JSON.Title)
			str += fmt.Sprintf("%s\n", v.JSON.Content)
		}
		return str, nil
	}
	ls := func(v interface{}) (string, error) {
		gs := isValidGS(v)
		var str string
		for key := range gs.Gmap {
			str += fmt.Sprintln(key)
		}
		return str, nil
	}

	sub := gomem.SubNewWithBase()
	sub.Addf("show",gs,  show, "show Gmap")
	sub.Addf("ls", gs, ls, "ls Gmap keys")
	// TODO: many addf
	//sub.Addf()
	if err := sub.Repl(r, w, "gomem:>"); err != nil {
		return err
	}
	return nil
}

// mock
func run(w io.Writer, gs *gomem.Gomems, opt *option) error {
	fmt.Fprintln(w, "debug: run")
	fmt.Fprintf(w, "debug: %v\n%v\n", gs, opt)

	return nil
}

func main() {
	if err := os.Chdir(opt.workdir); err != nil {
		log.Fatal(err)
	}
	gs, err := gomem.GomemsNew()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("debug:", gs)
	fmt.Println("debug:", opt)
	if opt.interactive {
		if err := interactive(os.Stdin, os.Stdout, gs); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := run(os.Stdout, gs, &opt); err != nil {
		log.Fatal(err)
	}
}
