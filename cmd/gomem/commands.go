package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
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

// mod path for json
func path2json(s *string) {
	if !strings.HasSuffix(*s, ".json") {
		*s = *s + ".json"
	}
}

/// commands ///
// status //
func la() (string, error) {
	var str string
	for key, v := range igs.Gmap {
		str += color.GreenString("----- %s -----\n", key)
		str += color.MagentaString("[ %s ]\n", v.J.Title)
		str += color.CyanString("%s\n", strings.Join(v.J.Content, "\n"))
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
				str += color.HiGreenString("sub categories:%s\n", info.Name())
			}
		}
	}
	for key, v := range igs.Gmap {
		str += color.GreenString("%s:", key)
		str += color.MagentaString("[ %s ]:", v.J.Title)
		str += fmt.Sprint("read only ")
		if v.Override {
			str += color.HiCyanString("%v\n", !v.Override)
		} else {
			str += color.RedString("%v\n", !v.Override)
		}
	}
	return str, nil
}
func show(s string) (string, error) {
	path2json(&s)
	g, ok := igs.Gmap[s]
	if !ok {
		return "not found:" + color.GreenString(s), nil
	}
	return color.CyanString("%s\n", strings.Join(g.J.Content, "\n")), nil
}
func todo() (string, error) {
	var str string
	var done string
	for key, g := range igs.Gmap {
		if strings.HasSuffix(g.J.Title, ":done") {
			done += color.GreenString("%s:", key)
			done += color.RedString("[ %s ]\n", g.J.Title)
			done += color.CyanString("\t%s\n", strings.Join(g.J.Content, "\n\t"))
			continue
		}
		if strings.HasPrefix(key, "todo"+string(filepath.Separator)) {
			str += color.GreenString("%s:", key)
			str += color.MagentaString("[ %s ]\n", g.J.Title)
			str += color.CyanString("\t%s\n\n", strings.Join(g.J.Content, "\n\t"))
		}
	}
	return done + str, nil
}

// contact to cache //
func newGomem() (string, error) {
	// trim ..
	fpath := filepath.Join(igs.GetDir(), read(prefname))
	path2json(&fpath)
	g, err := gomem.New(fpath, true)
	if err != nil {
		return err.Error(), nil
	}
	g.J.Title = read(pretitle)
	g.J.Content = append(g.J.Content, read(precontent))
	if err := igs.AddGomem(g); err != nil {
		return err.Error(), nil
	}
	return "new gomem key:" + color.GreenString(fpath), nil
}
func newGomemWithName(s string) (string, error) {
	s = filepath.Join(igs.GetDir(), path.Clean(s))
	path2json(&s)
	g, err := gomem.New(s, true)
	if err != nil {
		return err.Error(), nil
	}
	g.J.Title = read(pretitle)
	g.J.Content = append(g.J.Content, read(precontent))
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
	return "data cache reincluded: from " + color.HiGreenString(igs.GetDir()), nil
}
func cd() (string, error) {
	// TODO: cd: maybe don't needs use
	//     : consider delete cd()
	if !confirm("cd is dropped all data cache") {
		return "", nil
	}
	pwd := igs.GetDir()
	dir, err := filepath.Abs(filepath.Join(pwd, read("cd category:>")))
	if err != nil {
		return err.Error(), nil
	}
	// reconsider: needs it?
	if err := os.Chdir(dir); err != nil {
		return err.Error(), nil
	}
	tmpgs, err := gomem.GomemsNew(dir)
	if err != nil {
		// reconsider: needs it?
		if err := os.Chdir(pwd); err != nil {
			return err.Error(), nil
		}
		return err.Error(), nil
	}
	igs = tmpgs
	return "changed directory to:" + color.HiGreenString(igs.GetDir()), nil
}
func modContent(s string) (string, error) {
	path2json(&s)
	g, ok := igs.Gmap[s]
	if !ok {
		return "not found:" + s, nil
	}
	msg := color.GreenString("%s:", s) +
		color.MagentaString("[ %s ]", g.J.Title) +
		color.CyanString("%s\n", g.J.Content)
	c := read(msg + "mod " + precontent)
	g.J.Content = append(g.J.Content, c)
	return color.GreenString("content modified"), nil
}
func toggleReadonly(s string) (string, error) {
	path2json(&s)
	g, ok := igs.Gmap[s]
	if !ok {
		return "not found" + color.GreenString(s), nil
	}
	g.Override = !g.Override
	str := color.GreenString("key:%s", s)
	str += color.HiRedString("readonly:%+v", g.Override)
	return str, nil
}
func removeCache(s string) (string, error) {
	path2json(&s)
	if _, ok := igs.Gmap[s]; !ok {
		return "not found:" + color.GreenString(s), nil
	}
	if confirm("remove cache:"+s) == false {
		return "", nil
	}
	delete(igs.Gmap, s)
	return color.RedString("removed cache:" + s), nil
}
func appendTodo(s string) (string, error) {
	path2json(&s)
	s = filepath.Join("todo", s)
	g, ok := igs.Gmap[s]
	if !ok {
		return "not found:" + color.GreenString(s), nil
	}
	g.J.Content = append(g.J.Content, read("append "+precontent))
	g.J.Title = strings.TrimSuffix(g.J.Title, ":done")
	return "cache in:" +
			color.GreenString("%s:", s) +
			color.MagentaString("[ %s ]\n", g.J.Title) +
			color.CyanString("%s", strings.Join(g.J.Content, "\n")),
		nil
}
func done(s string) (string, error) {
	path2json(&s)
	s = filepath.Join("todo", s)
	g, ok := igs.Gmap[s]
	if !ok {
		return "not found " + s, nil
	}
	if strings.HasSuffix(g.J.Title, ":done") {
		return "already done:" + color.GreenString(s), nil
	}
	g.J.Title += ":done"
	return color.GreenString("%s:", s) +
			color.RedString("[ %s ]\n", g.J.Title) +
			color.CyanString("%s\n", strings.Join(g.J.Content, "\n")),
		nil
}
func trim(s string) (string, error) {
	path2json(&s)
	s = filepath.Join("todo", s)
	g, ok := igs.Gmap[s]
	if !ok {
		return "not found" + color.GreenString(s), nil
	}
	var msg string
	for i, s := range g.J.Content {
		msg += fmt.Sprintf("%d: %s\n", i+1, color.CyanString(s))
	}
	trimIndex, err := strconv.Atoi(read(msg + "line :> "))
	if err != nil {
		return err.Error(), nil
	}
	if trimIndex <= 0 || trimIndex > len(g.J.Content) {
		return "invalid line number:" + strconv.Itoa(trimIndex), nil
	}
	trimIndex--
	g.J.Content = append(g.J.Content[:trimIndex], g.J.Content[trimIndex+1:]...)
	return color.CyanString("%s", strings.Join(g.J.Content, "\n")), nil
}

// physical //
func makeSubcategory(s string) (string, error) {
	subname := filepath.Join(igs.GetDir(), filepath.Base(s))
	err := os.Mkdir(subname, 0777)
	if err != nil {
		return err.Error(), nil
	}
	return "maked subcategory:" + color.HiGreenString(subname), nil
}
func write() (string, error) {
	b := confirm("write all cache in " + color.HiGreenString(igs.GetDir()))
	var result string
	if b {
		for key, x := range igs.Gmap {
			if err := x.WriteFile(); err != nil {
				result += color.RedString("err:%s:%s\n", key, err.Error())
			}
		}
		return result, nil
	}
	return "stop write", nil
}
func remove(s string) (string, error) {
	path2json(&s)
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
		return "invalid category:" + color.HiGreenString(s), nil
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
	path2json(&s)
	s = filepath.Join("todo", s)
	g, err := gomem.New(filepath.Join(igs.GetDir(), s), true)
	if err != nil {
		return err.Error(), nil
	}
	t := time.Now()
	ts := fmt.Sprintf("%d %s %d %d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	g.J.Title = fmt.Sprintf("<%s>:%s", ts, s)
	g.J.Content = append(g.J.Content, read(precontent))
	igs.Gmap[s] = g
	return "cache in:" + color.GreenString("%s\n", s) +
			color.MagentaString("[ %s ]\n", g.J.Title) +
			color.CyanString("%s", strings.Join(g.J.Content, "\n")),
		nil
}

// interactive make interactive session
func interactive(r io.Reader, w io.Writer, prefix string, gs *gomem.Gomems, autoRuns []string, callBacks []string) error {
	if gs == nil || gs.Gmap == nil {
		return fmt.Errorf("gs or gs.Gmap is nil, exit session")
	}
	igs = gs
	interReader = r
	interWriter = w

	sub := gomem.SubNew(r, w)
	sub.Addf("exit", sub.Exit, "call exit")
	sub.Addf(":q", sub.Exit, "exit alias")
	sub.Addf("help", sub.Help, "show subcommands")
	sub.Addf("la", la, "show gs.Gmap")
	sub.Addf("ls", ls, "ls gs.Gmap keys")
	sub.Addf("state", state, "show state of gs")
	sub.Addf("new", newGomem, "new gomem")
	sub.Addf("write", write, "write all data to gs.dir")
	sub.Addf("cd", cd, "change working directory, and exchange of data cache")
	sub.Addf("todo", todo, "subcategory [todo/*]")
	sub.Addf("include", include, "reinclude from gs.dir")

	sub.Addfa("show", show, "show title and content")
	sub.Addfa("mkdir", makeSubcategory, "mkdir make subcategory in gs.dir")
	sub.Addfa("rm", remove, "remove physical file")
	sub.Addfa("rmsub", removeSubcategory, "remove subcategory directory")
	sub.Addfa("rmcache", removeCache, "remove cache data")
	sub.Addfa("new", newGomemWithName, "")
	sub.Addfa("mod", modContent, "modify content")
	sub.Addfa("todo", createTodo, "")
	sub.Addfa("done", done, "for [todo/*] check done flag")
	sub.Addfa("append", appendTodo, "append todo")
	sub.Addfa("trim", trim, "trim in todo")
	sub.Addfa("readonly!", toggleReadonly, "toggle readonly falg")

	if autoRuns != nil {
		sub.InterCh = make(chan string, len(autoRuns))
		for _, s := range autoRuns {
			sub.InterCh <- s
		}
	}
	if callBacks != nil {
		sub.CallBackBuf = make(chan string, len(callBacks))
		for _, s := range callBacks {
			sub.CallBackBuf <- s
		}
	}
	sub.Prefix = prefix
	if err := sub.Repl(); err != nil {
		return err
	}
	return nil
}
