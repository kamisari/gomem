package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/kamisari/gomem"
)

const version = "0.0"

type option struct {
	version bool
	workdir string

	// TODO: impl subcmd for run
	subcmd  string
	subarsg []string
}

var opt option

func (opt *option) init() {
	flag.BoolVar(&opt.version, "version", false, "")
	flag.BoolVar(&opt.version, "v", false, "")

	flag.StringVar(&opt.workdir, "workdir", "", "")
	flag.StringVar(&opt.workdir, "w", "", "")
	flag.Parse()
	if flag.NArg() != 0 {
		// TODO: impl parse subcmd
		log.Fatalf("invalid args: %q", flag.Args())
	}
	if opt.version {
		fmt.Printf("version %s\n", version)
		os.Exit(0)
	}
	if opt.workdir == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		opt.workdir = filepath.Join(u.HomeDir, "dotfiles", "etc", "gomem")
	}
	var err error
	opt.workdir, err = filepath.Abs(opt.workdir)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Chdir(opt.workdir); err != nil {
		log.Fatal(err)
	}
}

func init() {
	log.SetPrefix("gomem:")
	log.SetFlags(log.Lshortfile)

	opt.init()
}

func main() {
	gs, err := gomem.GomemsNew()
	if err != nil {
		log.Fatal(err)
	}
	if opt.subcmd != "" {
		if err := run(os.Stdout, gs); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := interactive(os.Stdin, os.Stdout, "gomem:>", gs); err != nil {
		log.Fatal(err)
	}
}

// mock TODO: impl
// specify not interactive then run the this function
func run(w io.Writer, gs *gomem.Gomems) error {
	fmt.Fprintln(w, "debug: run")
	return nil
}
