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

// Gomem JSON structure
type Gomem struct {
	Title   string `json:"title"`
	Content string `json:"content"`

	base     string
	override bool
}

// Gomems map of Gomem and data directory
type Gomems struct {
	Gmap map[string]*Gomem // key: base(filepath)
	dir  string // reconsider base directory
}

// ErrFileExists exists error
var ErrFileExists = errors.New("file exists, cannot override")

// WritePerm if need then modify
// Example: gomem.WritePerm = os.FileMode(0666)
var WritePerm = os.FileMode(0600)

// New return *Gomem init filepath and override flag
// if fpath is exists and write is flase then return error
// accepted filename is "*.json" only
func New(fpath string, override bool) (*Gomem, error) {
	if !strings.HasSuffix(fpath, ".json") {
		return nil, fmt.Errorf("invalid filename:%s require filename is *.json", fpath)
	}
	return &Gomem{base: filepath.Base(fpath), override: override}, nil
}

// ModOverride is modify override flag
func (g *Gomem) ModOverride(b bool) {
	g.override = b
}

// IsOverride return override boolean
func (g *Gomem) IsOverride() bool {
	return g.override
}

// IsValidFilePath if invalid then return error
// verification file path
// for write function
func (g *Gomem) IsValidFilePath() error {
	if !strings.HasSuffix(g.base, ".json") {
		return fmt.Errorf("*Gomem.IsValidFilePath:%s: require file name *.json", g.base)
	}
	if info, err := os.Stat(g.base); os.IsNotExist(err) {
		return nil
	} else if err == nil && info.Mode().IsRegular() {
		return nil
	}
	return fmt.Errorf("*Gomem.IsValidFilePath: invalid filename:%s", g.base)
}

// ReadFile load from g.base
func (g *Gomem) ReadFile() error {
	b, err := ioutil.ReadFile(g.base)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, g)
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
	if _, err := os.Stat(g.base); err == nil && g.IsOverride() != true {
		return ErrFileExists
	}
	b, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(g.base, b, WritePerm); err != nil {
		return err
	}
	return nil
}

// GomemsNew read from dir return map for Gomem
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
	if gs.Gmap == nil {
		return fmt.Errorf("*Gomems.AddGomem: gs.Gmap == nil")
	}
	if _, ok := gs.Gmap[g.base]; ok {
		return fmt.Errorf("*Gomems.AddGomem: gs.Gmap[%s] is exists", g.base)
	}
	gs.Gmap[g.base] = g
	return nil
}

// IncludeJSON include from Gomems.dir
// mapping gs.Gmap[g.base]*g
func (gs *Gomems) IncludeJSON() error {
	if gs.Gmap == nil {
		return fmt.Errorf("*Gomems.IncludeJSON: Gmap is nil")
	}
	bases, err := getBaseJSON(gs.dir)
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
		// TODO: reconsider boolean
		g, err := New(x, false)
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

// return json files base name in dir
func getBaseJSON(dir string) ([]string, error) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var str []string
	for _, info := range infos {
		if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".json") {
			str = append(str, info.Name())
		}
	}
	return str, nil
}
