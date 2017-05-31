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

// Gomem JSON structure
type Gomem struct {
	J        JSON
	Override bool
	// TODO: to fullpath
	base     string
}

// Gomems map of Gomem and data directory
type Gomems struct {
	Gmap map[string]*Gomem // key: filepath
	dir  string            // reconsider export
}

// ErrFileExists exists error
var ErrFileExists = errors.New("file exists, cannot override")

// WritePerm if need then modify
// This use *Gomem.WriteFile
// Example: gomem.WritePerm = os.FileMode(0666)
var WritePerm = os.FileMode(0600)

// New return *Gomem
// fpath is modify to clean name
// accepted filename is "*.json" only
func New(fpath string, override bool) (*Gomem, error) {
	if !strings.HasSuffix(fpath, ".json") {
		return nil, fmt.Errorf("New: invalid filename:%s require filename is *.json", fpath)
	}
	return &Gomem{base: filepath.Clean(fpath), Override: override}, nil
}

// IsValidFilePath if invalid then return error
// verification file path
// for *Gomem.WriteFile
func (g *Gomem) IsValidFilePath() error {
	if !strings.HasSuffix(g.base, ".json") {
		return fmt.Errorf("*Gomem.IsValidFilePath:%s: require file name *.json", g.base)
	}
	if info, err := os.Stat(g.base); os.IsNotExist(err) {
		return nil
	} else if err == nil && info.Mode().IsRegular() {
		return nil
	}
	return fmt.Errorf("*Gomem.IsValidFilePath: invalid filename maybe is not regular files:%s", g.base)
}

// ReadFile load from g.base
func (g *Gomem) ReadFile() error {
	b, err := ioutil.ReadFile(g.base)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &g.J)
	if err != nil {
		return err
	}
	return nil
}

// WriteFile write to g.base
func (g *Gomem) WriteFile() error {
	if err := g.IsValidFilePath(); err != nil {
		return err
	}
	// reconsider dupl check
	if _, err := os.Stat(g.base); err == nil && g.Override != true {
		return ErrFileExists
	}
	b, err := json.MarshalIndent(g.J, "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(g.base, b, WritePerm); err != nil {
		return err
	}
	return nil
}

// GomemsNew read from pwd return map for Gomem
func GomemsNew() (*Gomems, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	gs := &Gomems{
		Gmap: make(map[string]*Gomem),
		dir:  pwd,
	}
	if err := gs.IncludeJSON(); err != nil {
		return nil, err
	}
	return gs, nil
}

// AddGomem add to gs.Gmap
func (gs *Gomems) AddGomem(g *Gomem) error {
	if _, ok := gs.Gmap[g.base]; ok {
		return fmt.Errorf("*Gomems.AddGomem: gs.Gmap[%s] is exists", g.base)
	}
	gs.Gmap[g.base] = g
	return nil
}

// GetAbs return filepath.Join(gs.dir+gs.Gmap[key].base)
func (gs *Gomems) GetAbs(key string) (string, error) {
	g, ok := gs.Gmap[key]
	if ok {
		return filepath.Join(gs.dir, g.base), nil
	}
	return "", fmt.Errorf("not found gs.Gmap[%s]", key)
}

// IncludeJSON include from Gomems.dir
// mapping gs.Gmap[g.base]*g
func (gs *Gomems) IncludeJSON() error {
	if gs.Gmap == nil {
		return fmt.Errorf("*Gomems.IncludeJSON: Gmap is nil")
	}
	var bases []string
	var pushBases func(string) error
	pushBases = func(root string) error {
		infos, err := ioutil.ReadDir(root)
		if err != nil {
			return err
		}
		for _, info := range infos {
			if info.IsDir() {
				return pushBases(filepath.Join(gs.dir, info.Name()))
			}
			if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".json") {
				path, err := filepath.Rel(gs.dir, filepath.Join(root, info.Name()))
				if err == nil {
					bases = append(bases, path)
				}
			}
		}
		return nil
	}
	err := pushBases(gs.dir)
	if err != nil {
		return err
	}
	for _, x := range bases {
		if g, ok := gs.Gmap[x]; ok {
			if err := g.ReadFile(); err != nil {
				return err
			}
			continue
		}
		g, err := New(x, true)
		if err != nil {
			return err
		}
		gs.Gmap[x] = g
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
