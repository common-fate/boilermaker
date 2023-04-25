package boilermaker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

type Metadata struct {
	Description string `json:"description"`
}

type Boilerplate struct {
	Template *template.Template
	Metadata Metadata
}

func (b *Boilerplate) ParseMetadata(fsys fs.FS, path string) error {
	f, err := fsys.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(&b.Metadata)
}

func ParseFS(fsys fs.FS) (*Boilerplate, error) {
	bp := Boilerplate{
		Template: template.New(""),
	}
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "_boilermaker.json" {
			return bp.ParseMetadata(fsys, path)
		}

		// skip dirs
		if d.IsDir() {
			return nil
		}

		f, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		fileBytes, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		bp.Template.New(path).Parse(string(fileBytes))

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &bp, nil
}

// Generate a map of files for the boilerplate.
func (bp Boilerplate) Generate(data map[string]any) (Result, error) {
	output := Result{}

	for _, tmpl := range bp.Template.Templates() {
		var b bytes.Buffer
		err := tmpl.Execute(&b, data)
		if err != nil {
			return nil, err
		}
		// template the actual name of the file too
		nameTmpl, err := template.New("").Parse(tmpl.Name())
		if err != nil {
			return nil, err
		}
		var name bytes.Buffer
		err = nameTmpl.Execute(&name, data)
		if err != nil {
			return nil, err
		}
		output[name.String()] = b.String()
	}
	return output, nil
}

// Result is a map of filenames to file contents.
type Result map[string]string

// Write the generated files to disk
func (r Result) Write(dir string) error {
	for f, contents := range r {
		fullpath := filepath.Join(dir, f)
		parent := filepath.Dir(fullpath)
		err := os.MkdirAll(parent, 0755)
		if err != nil {
			return err
		}

		err = os.WriteFile(fullpath, []byte(contents), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseMapFS parses a set of boilerplate directories.
// it expects that the file structure is as follows:
//   - one
//   - one/example
//   - two
//   - two/something_else
//
// and these will be parsed as boilerplates 'one' and 'two'
func ParseMapFS(fsys fs.FS, dir string) (map[string]*Boilerplate, error) {
	subfolders, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}

	set := map[string]*Boilerplate{}

	for _, subfolder := range subfolders {
		if !subfolder.IsDir() {
			return nil, fmt.Errorf("%s was not a folder", subfolder.Name())
		}
		fullpath := filepath.Join(dir, subfolder.Name())
		nested, err := fs.Sub(fsys, fullpath)
		if err != nil {
			return nil, err
		}
		bp, err := ParseFS(nested)
		if err != nil {
			return nil, err
		}
		set[subfolder.Name()] = bp
	}

	return set, nil
}
