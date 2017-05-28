package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/kamisari/gomem"
)

// for Read and confirm
// accept exchange output and input
// default: writer = os.Stdout
//        : reader = os.Stdin
var (
	igs         *gomem.Gomems
	interWriter io.Writer = os.Stdout
	interReader io.Reader = os.Stdin
)

var (
	prefname   = color.GreenString("filename:> ")
	pretitle   = color.MagentaString("title:> ")
	precontent = color.CyanString("content:> ")
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

/// commands ///
// status //
func la() (string, error) {
	var str string
	for key, v := range igs.Gmap {
		str += color.GreenString("----- %s -----\n", key)
		str += color.MagentaString("[ %s ]\n", v.JSON.Title)
		str += color.CyanString("%s\n", v.JSON.Content)
	}
	return str, nil
}
func ls() (string, error) {
	var str string
	for key := range igs.Gmap {
		str += color.GreenString("%s\n", key)
	}
	return str, nil
}
func state() (string, error) {
	var str string
	str += color.GreenString("igs.dir:%s\n", igs.GetDir())
	infos, err := ioutil.ReadDir(igs.GetDir())
	if err == nil {
		for _, info := range infos {
			if info.IsDir() {
				str += color.HiGreenString("sub category:%s\n", info.Name())
			}
		}
	}
	for key, v := range igs.Gmap {
		str += color.GreenString("%s:", key)
		str += color.MagentaString("[ %s ]:", v.JSON.Title)
		str += fmt.Sprintln("read only", !v.Override)
	}
	return str, nil
}
func show(s string) (string, error) {
	if !strings.HasSuffix(s, ".json") {
		s = s + ".json"
	}
	g, ok := igs.Gmap[s]
	if !ok {
		return "not found:" + s, nil
	}
	str := color.MagentaString("[ %s ]\n", g.JSON.Title)
	str += color.CyanString(g.JSON.Content)
	return str, nil
}

// contact to cache //
func newGomem() (string, error) {
	fpath := read(prefname)
	g, err := gomem.New(fpath, true)
	if err != nil {
		return err.Error(), nil
	}
	g.JSON.Title = read(pretitle)
	g.JSON.Content = read(precontent)
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
	g.JSON.Title = read(pretitle)
	g.JSON.Content = read(precontent)
	if err := igs.AddGomem(g); err != nil {
		return err.Error(), nil
	}
	return "new gomem included", nil
}
func include() (string, error) {
	err := igs.IncludeJSON()
	if err != nil {
		return err.Error(), nil
	}
	return "data cache reincluded: from " + igs.GetDir(), nil
}
func cd() (string, error) {
	if !confirm("cd is dropped all data cache") {
		return "", nil
	}
	pwd := igs.GetDir()
	if err := os.Chdir(read("cd category:>")); err != nil {
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
	return "changed directory to:" + color.HiGreenString(igs.GetDir()), nil
}
func modContent(s string) (string, error) {
	if !strings.HasSuffix(s, ".json") {
		s = s + ".json"
	}
	g, ok := igs.Gmap[s]
	if !ok {
		return "not found:" + s, nil
	}
	msg := color.GreenString("%s:", s) + color.MagentaString("[ %s ]", g.JSON.Title) + color.CyanString("%s\n", g.JSON.Content)
	c := read(msg + "mod " + precontent)
	g.JSON.Content = c
	return color.GreenString("content modified"), nil
}
func removeCache(s string) (string, error) {
	if !strings.HasSuffix(s, ".json") {
		s = s + ".json"
	}
	if _, ok := igs.Gmap[s]; !ok {
		return "not found:" + s, nil
	}
	if confirm("remove cache:"+s) == false {
		return "", nil
	}
	delete(igs.Gmap, s)
	return color.RedString("removed cache:" + s), nil
}
func todo() (string, error) {
	var str string
	for key, g := range igs.Gmap {
		if strings.HasPrefix(key, "todo"+string(filepath.Separator)) {
			str += color.GreenString("%s:", key)
			if strings.HasSuffix(g.JSON.Title, ":done") {
				str += color.RedString("[ %s ]\n", g.JSON.Title)
			} else {
				str += color.MagentaString("[ %s ]\n", g.JSON.Title)
			}
			str += color.CyanString("\t%s\n", g.JSON.Content)
		}
	}
	return str, nil
}

// physical //
func makeSubcategory(s string) (string, error) {
	subname := filepath.Join(igs.GetDir(), filepath.Base(s))
	err := os.Mkdir(subname, 0777)
	if err != nil {
		return err.Error(), nil
	}
	return "maked subcategory:" + subname, nil
}
func write() (string, error) {
	b := confirm("write all cache in " + igs.GetDir())
	var result string
	if b {
		for key, x := range igs.Gmap {
			if err := x.WriteFile(); err != nil {
				result += color.RedString("err:%s:%s\n", key, err.Error())
			}
		}
		return result, nil
	}
	return "", nil
}
func remove(s string) (string, error) {
	if !strings.HasSuffix(s, ".json") {
		s = s + ".json"
	}
	fullpath, err := igs.GetAbs(s)
	if err != nil {
		return err.Error(), nil
	}
	if confirm("remove:"+fullpath) == false {
		return "", nil
	}
	err = os.Remove(fullpath)
	if err != nil {
		return err.Error(), nil
	}
	delete(igs.Gmap, s)
	return color.RedString(fullpath + " is removed"), nil
}
func removeSubcategory(s string) (string, error) {
	subname := filepath.Join(igs.GetDir(), filepath.Base(s))
	info, err := os.Lstat(subname)
	if err != nil {
		return err.Error(), nil
	}
	if !info.IsDir() {
		return "invalid category:" + s, nil
	}
	if !confirm("remove all files in " + subname) {
		return "", nil
	}
	err = os.RemoveAll(subname)
	if err != nil {
		return err.Error(), nil
	}
	return color.RedString("removed subcategory:" + subname), nil
}
func createTodo(s string) (string, error) {
	if !strings.HasSuffix(s, ".json") {
		s = s + ".json"
	}
	s = filepath.Join("todo", s)
	g, ok := igs.Gmap[s]
	if !ok {
		var err error
		g, err = gomem.New(s, true)
		if err != nil {
			return err.Error(), nil
		}
	}
	t := time.Now()
	ts := fmt.Sprintf("%d %s %d %d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	g.JSON.Title = fmt.Sprintf("<%s>:%s", ts, s)
	g.JSON.Content = read(precontent)
	igs.Gmap[s] = g
	return "cache in:" + color.GreenString("%s:", s) + color.MagentaString("[ %s ]:", g.JSON.Title) + color.CyanString("%s", g.JSON.Content), nil
}
func done(s string) (string, error) {
	if !strings.HasSuffix(s, ".json") {
		s = s + ".json"
	}
	s = filepath.Join("todo", s)
	g, ok := igs.Gmap[s]
	if !ok {
		return "not found" + s, nil
	}
	if strings.HasSuffix(g.JSON.Title, ":done") {
		return "already done:" + color.GreenString(s), nil
	}
	g.JSON.Title += ":done"
	return color.GreenString("%s:", s) + color.RedString("[ %s ]:", g.JSON.Title) + color.CyanString("%s", g.JSON.Content), nil
}

// interactive make interactive session
func interactive(r io.Reader, w io.Writer, prefix string, gs *gomem.Gomems, firstRun string, autoWrite bool) error {
	if gs == nil || gs.Gmap == nil {
		return fmt.Errorf("gs or gs.Gmap is nil, exit session")
	}
	igs = gs
	interReader = r
	interWriter = w

	sub := gomem.SubNewWithBase(r, w)
	sub.Addf("la", la, "show gs.Gmap")
	sub.Addf("ls", ls, "ls gs.Gmap keys")
	sub.Addf(":q", gomem.Exit, "exit alias")
	sub.Addf("state", state, "show state of gs")
	sub.Addf("new", newGomem, "new gomem")
	sub.Addf("write", write, "write all data to gs.dir")
	sub.Addf("cd", cd, "change working directory, and exchange of data cache")
	sub.Addf("todo", todo, "subcategory [todo/*]")
	sub.Addf("include", include, "reinclude from gs.dir")

	sub.Addfa("show", show, "show title and content")
	sub.Addfa("mkdir", makeSubcategory, "mkdir make subcategory in gs.dir")
	sub.Addfa("remove", remove, "remove json file")
	sub.Addfa("removesub", removeSubcategory, "remove subcategory directory")
	sub.Addfa("removecache", removeCache, "remove cache data")
	sub.Addfa("new", newGomemWithName, "")
	sub.Addfa("mod", modContent, "modify content")
	sub.Addfa("todo", createTodo, "")
	sub.Addfa("done", done, "for [todo/*] check done flag")

	if firstRun != "" {
		sub.InterCh <- firstRun
	}
	if autoWrite {
		sub.AddCallBack(func() {
			result, err := write()
			if err != nil {
				panic(err)
			}
			fmt.Fprintln(interWriter, result)
		})
	}
	if err := sub.Repl(prefix); err != nil {
		return err
	}
	return nil
}
