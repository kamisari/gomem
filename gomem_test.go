package gomem

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var (
	tmpdir  string
	tmpfile string
)

func TestMain(m *testing.M) {
	var err error
	tmpdir, err = ioutil.TempDir("t", "gomemtest")
	if err != nil {
		log.Fatal(err)
	}
	tmpfile = filepath.Join(tmpdir, "file.json")
	if _, err = os.Create(tmpfile); err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()
	defer os.Exit(exitCode)

	os.RemoveAll(tmpdir)
}

func TestNew(t *testing.T) {
	type input struct {
		path string
		flag bool
	}
	var gomemTests = []struct {
		in      input
		expBase string
		wantErr bool
	}{
		// invalid in
		{
			in:      input{path: "", flag: false},
			wantErr: true,
		},
		{
			in:      input{path: "./test.go", flag: false},
			wantErr: true,
		},

		// valid in
		{
			in:      input{path: "./foo.json", flag: false},
			expBase: "foo.json",
			wantErr: false,
		},
		{
			in:      input{path: "/home/json/test/json.json"},
			expBase: "json.json",
			wantErr: false,
		},
	}
	for _, v := range gomemTests {
		g, err := New(v.in.path, v.in.flag)
		if v.wantErr && err != nil {
			continue
		}
		if err != nil {
			t.Error(err)
			continue
		}
		if g == nil {
			t.Errorf("failed initalize: g == nil")
			continue
		}
		if v.expBase != g.base {
			t.Errorf("exp: %s\nout:%s", v.expBase, g.base)
		}
	}
}

func TestGomem_IsValidFilePath(t *testing.T) {
	dirname := filepath.Join(tmpdir, "dir.json")
	if err := os.Mkdir(dirname, os.ModeDir|0700); err != nil {
		t.Fatal(err)
	}
	filename := tmpfile
	tests := []struct {
		g       *Gomem
		wantErr bool
	}{
		// invalid
		{g: &Gomem{base: ""}, wantErr: true},
		{g: &Gomem{base: "/path/to/file.go"}, wantErr: true},
		{g: &Gomem{base: dirname}, wantErr: true}, // dir name
		// valid
		{g: &Gomem{base: "file.json"}, wantErr: false},
		{g: &Gomem{base: filename}, wantErr: false},
	}
	for _, v := range tests {
		err := v.g.IsValidFilePath()
		if v.wantErr && err != nil {
			continue
		}
		if err != nil {
			t.Errorf("g.base:%s, err:%v", v.g.base, err)
			continue
		}
	}
}

func TestGomem_ReadFile(t *testing.T) {
	invalidFilePath, err := ioutil.TempFile(tmpdir, "invalid")
	if err != nil {
		t.Fatal(err)
	}
	defer invalidFilePath.Close()
	if err = os.Remove(invalidFilePath.Name()); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		fullPath string
		data     []byte
		wantErr  bool
	}{
		{
			fullPath: tmpfile,
			data:     []byte(`{"title": "test", "content": "test"}`),
			wantErr:  false,
		},
		{
			fullPath: invalidFilePath.Name(),
			wantErr:  true,
		},
		{
			fullPath: tmpfile,
			data:     []byte("invalid content"),
			wantErr:  true,
		},
	}
	for _, v := range tests {
		if err := ioutil.WriteFile(tmpfile, v.data, 0666); err != nil {
			t.Fatal(err)
		}
		g := &Gomem{base: v.fullPath}
		err := g.ReadFile()
		if v.wantErr && err != nil {
			continue
		} else if err != nil {
			t.Error(err)
		}
	}
}

func TestGomem_WriteFile(t *testing.T) {
	// reconsider
	tests := []*struct {
		in  gomemJSON
		out string
	}{
		{in: gomemJSON{Title: "test", Content: []string{"test"}}},
	}
	for _, v := range tests {
		v.out = fmt.Sprintf("{\n  \"title\": %q,\n  \"content\": %q\n}", v.in.Title, v.in.Content)
	}

	for _, v := range tests {
		g := &Gomem{base: tmpfile, JSON: v.in, Override: true}
		if err := g.WriteFile(); err != nil {
			t.Fatal(err)
		}
		b, err := ioutil.ReadFile(g.base)
		if err != nil {
			t.Fatal(err)
		}
		if v.out != string(b) {
			t.Errorf("writed:%q, but expected:%q", string(b), v.out)
		}
	}
}

func TestGomemsNew(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(pwd)
	// os.Chdir(to test directory)

	// TODO: impl
	_, err = GomemsNew()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGomems_AddGomem(t *testing.T) {
	//	t.Fatal()
}
func TestGomems_IncludeJSON(t *testing.T) {
	//	t.Fatal()
}
func TestGomems_GetDir(t *testing.T) {
	//	t.Fatal()
}
