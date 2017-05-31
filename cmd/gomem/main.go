package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/kamisari/gomem"
)

const version = "0.0.0"

type option struct {
	version     bool
	workdir     string
	autocmd     string
	autowrite   bool
	interactive bool
}

var opt option

func (opt *option) init() {
	flag.BoolVar(&opt.version, "version", false, "")
	flag.StringVar(&opt.workdir, "workdir", "", "")
	flag.StringVar(&opt.autocmd, "autocmd", "todo", "")
	flag.BoolVar(&opt.autowrite, "autowrite", false, "")
	flag.BoolVar(&opt.autowrite, "w", false, "alias of autowrite")
	flag.BoolVar(&opt.interactive, "interactive", false, "")
	flag.BoolVar(&opt.interactive, "i", false, "alias of interactive")
	flag.Parse()
	if flag.NArg() != 0 {
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
	if opt.interactive == false {
		opt.autocmd = opt.autocmd + string(filepath.ListSeparator) + "exit"
	}
}

func init() {
	log.SetPrefix("gomem:")
	log.SetOutput(os.Stderr)
	opt.init()
}

func main() {
	gs, err := gomem.GomemsNew()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("autocmd:", filepath.SplitList(opt.autocmd))
	if err := interactive(os.Stdin, os.Stdout, "gomem:> ", gs, filepath.SplitList(opt.autocmd), opt.autowrite); err != nil {
		log.Fatal(err)
	}
}
