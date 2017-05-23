package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/kamisari/gomem"
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
		if err := gomem.Interactive(os.Stdin, os.Stdout, "gomem:>", gs); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := run(os.Stdout, gs, &opt); err != nil {
		log.Fatal(err)
	}
}
