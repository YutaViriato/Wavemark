// Package scaffold creates the starter site written by `wavemark init`.
package scaffold

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

//go:embed starter
var starter embed.FS

// Init writes the starter site into dir. It never overwrites a file: if any
// starter file already exists, nothing is written. It returns the paths it
// created, relative to dir.
func Init(dir string) ([]string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	title := titleFrom(filepath.Base(abs))

	var files []string
	err = fs.WalkDir(starter, "starter", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		files = append(files, strings.TrimPrefix(p, "starter/"))
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(abs, f)); err == nil {
			return nil, fmt.Errorf("%s already exists; init only writes into a folder without a site", filepath.Join(dir, f))
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	for _, f := range files {
		data, err := starter.ReadFile("starter/" + f)
		if err != nil {
			return nil, err
		}
		data = []byte(strings.ReplaceAll(string(data), "{{TITLE}}", yamlSafe(title)))
		target := filepath.Join(abs, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return nil, err
		}
	}
	return files, nil
}

// titleFrom turns a folder name into a title: "my-docs" → "My docs".
func titleFrom(name string) string {
	s := strings.TrimSpace(strings.NewReplacer("-", " ", "_", " ", ".", " ").Replace(name))
	if s == "" || s == "/" {
		return "My site"
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// yamlSafe keeps the title valid both in YAML and in a quoted attribute.
func yamlSafe(s string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(`:"'#{}[]<>&*!|>%@`+"`", r) {
			return -1
		}
		return r
	}, s)
}
