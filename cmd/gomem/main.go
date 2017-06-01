package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/kamisari/gomem"
)

const version = "0.0.0"

// TODO: reconsider names
type option struct {
	version     bool
	workdir     string
	autocmd     string
	callback    string
	interactive bool
	conf        string
}

var opt option

func (opt *option) getAutoRunList() []string {
	var list []string
	if opt.conf != "" {
		b, err := ioutil.ReadFile(opt.conf)
		if err != nil {
			return nil
		}
		for _, s := range strings.Fields(string(b)) {
			if strings.HasPrefix(s, "autocmd=") {
				list = append(list, strings.TrimPrefix(s, "autocmd="))
			}
		}
	}
	if opt.autocmd != "" {
		list = append(list, strings.Fields(opt.autocmd)...)
	}
	if opt.interactive == false {
		list = append(list, "exit")
	}
	return list
}
func (opt *option) getCallbacks() []string {
	return strings.Fields(opt.callback)
}

// TODO: be graceful
func (opt *option) init() error {
	flag.BoolVar(&opt.version, "version", false, "")
	flag.StringVar(&opt.workdir, "workdir", "", "")
	flag.StringVar(&opt.autocmd, "autocmd", "todo", "")
	flag.StringVar(&opt.callback, "callback", "", "")
	flag.BoolVar(&opt.interactive, "interactive", false, "")
	flag.BoolVar(&opt.interactive, "i", false, "alias of interactive")
	flag.StringVar(&opt.conf, "conf", "", "path to configuration file")
	flag.Parse()
	if flag.NArg() != 0 {
		return fmt.Errorf("invalid args: %q", flag.Args())
	}
	if opt.version {
		fmt.Printf("version %s\n", version)
		os.Exit(0)
	}
	// default work directory
	if opt.workdir == "" {
		u, err := user.Current()
		if err != nil {
			return err
		}
		opt.workdir = filepath.Join(u.HomeDir, "dotfiles", "etc", "gomem")
	}
	var err error
	opt.workdir, err = filepath.Abs(opt.workdir)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	log.SetPrefix("gomem:")
	log.SetOutput(os.Stderr)
	if err := opt.init(); err != nil {
		log.Fatal(err)
	}
	// reconsider: needs it?
	if err := os.Chdir(opt.workdir); err != nil {
		log.Fatal(err)
	}

	gs, err := gomem.GomemsNew(opt.workdir)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("autocmd:", opt.getAutoRunList())
	err = interactive(os.Stdin, os.Stdout, "gomem:> ", gs, opt.getAutoRunList(), opt.getCallbacks())
	if err != nil {
		log.Fatal(err)
	}
}
