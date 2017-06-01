package gomem

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// JSON JSON structure
type JSON struct {
	Title   string   `json:"title"`
	Content []string `json:"content"`
}

// Gomem have JSON structure
type Gomem struct {
	J        JSON
	Override bool
	fullpath string
}

// Gomems map of Gomem and data directory
type Gomems struct {
	Gmap map[string]*Gomem // key: filepath.Rel(Gomems.dir, Gomem.fullpath)
	dir  string
}

// ErrFileExists exists error
var ErrFileExists = errors.New("file exists, cannot override")

// WritePerm if need then modify
// This use *Gomem.WriteFile
// Example: gomem.WritePerm = os.FileMode(0600)
var WritePerm = os.FileMode(0666)

// New return *Gomem
// if fpath is not Abs then return error
// accepted filename is "*.json" only
func New(fpath string, override bool) (*Gomem, error) {
	if !strings.HasSuffix(fpath, ".json") {
		return nil, fmt.Errorf("New: invalid filename:%s require filename is *.json", fpath)
	}
	if !filepath.IsAbs(fpath) {
		return nil, fmt.Errorf("invalid filepath: %v is not fullpath", fpath)
	}
	return &Gomem{fullpath: fpath, Override: override}, nil
}

// IsValidFilePath if invalid then return error
// verification file path
// for *Gomem.WriteFile
func (g *Gomem) IsValidFilePath() error {
	if !strings.HasSuffix(g.fullpath, ".json") || !filepath.IsAbs(g.fullpath) {
		return fmt.Errorf("*Gomem.IsValidFilePath:%s: require file name *.json and fullpath", g.fullpath)
	}
	if info, err := os.Stat(g.fullpath); os.IsNotExist(err) {
		return nil
	} else if err == nil && info.Mode().IsRegular() {
		return nil
	}
	return fmt.Errorf("*Gomem.IsValidFilePath: invalid filename maybe is not regular files:%s", g.fullpath)
}

// ReadFile load from g.fullpath
func (g *Gomem) ReadFile() error {
	b, err := ioutil.ReadFile(g.fullpath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &g.J)
	if err != nil {
		return err
	}
	return nil
}

// WriteFile write to g.fullpath
func (g *Gomem) WriteFile() error {
	if err := g.IsValidFilePath(); err != nil {
		return err
	}
	// reconsider dupl check
	if _, err := os.Stat(g.fullpath); err == nil && g.Override != true {
		return ErrFileExists
	}
	b, err := json.MarshalIndent(g.J, "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(g.fullpath, b, WritePerm); err != nil {
		return err
	}
	return nil
}

// GomemsNew read from pwd return map for Gomem
func GomemsNew(dir string) (*Gomems, error) {
	if !filepath.IsAbs(dir) {
		return nil, fmt.Errorf("GomemsNew: invalid direcotry path %s", dir)
	}
	gs := &Gomems{
		Gmap: make(map[string]*Gomem),
		dir:  dir,
	}
	if err := gs.IncludeJSON(); err != nil {
		return nil, err
	}
	return gs, nil
}

// AddGomem add to gs.Gmap
func (gs *Gomems) AddGomem(g *Gomem) error {
	key, err := filepath.Rel(gs.dir, g.fullpath)
	if err != nil {
		return err
	}
	if _, ok := gs.Gmap[key]; ok {
		return fmt.Errorf("*Gomems.AddGomem: gs.Gmap[%s] is exists", key)
	}
	gs.Gmap[key] = g
	return nil
}

// GetAbs return filepath.Join(gs.dir+gs.Gmap[key].base)
func (gs *Gomems) GetAbs(key string) (string, error) {
	g, ok := gs.Gmap[key]
	if ok {
		return g.fullpath, nil
	}
	return "", fmt.Errorf("not found gs.Gmap[%s]", key)
}

// IncludeJSON include from Gomems.dir
// mapping gs.Gmap[g.base]*g
func (gs *Gomems) IncludeJSON() error {
	if gs.Gmap == nil {
		return fmt.Errorf("*Gomems.IncludeJSON: Gmap is nil")
	}

	var fullpaths []string
	var pushPaths func(string) error
	pushPaths = func(root string) error {
		infos, err := ioutil.ReadDir(root)
		if err != nil {
			return err
		}
		for _, info := range infos {
			if info.IsDir() {
				return pushPaths(filepath.Join(root, info.Name()))
			}
			if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".json") {
				if err == nil {
					fullpaths = append(fullpaths, filepath.Join(root, info.Name()))
				}
			}
		}
		return nil
	}
	err := pushPaths(gs.dir)
	if err != nil {
		return err
	}

	for _, x := range fullpaths {
		key, err := filepath.Rel(gs.dir, x)
		if err != nil {
			// TODO: to continue?
			return err
		}
		if g, ok := gs.Gmap[key]; ok {
			if err := g.ReadFile(); err != nil {
				return err
			}
			continue
		}
		g, err := New(x, true)
		if err != nil {
			return err
		}
		gs.Gmap[key] = g
		if err := g.ReadFile(); err != nil {
			return err
		}
	}
	return nil
}

// GetDir exported gs.dir
func (gs *Gomems) GetDir() string {
	return gs.dir
}
