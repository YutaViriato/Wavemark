package scaffold

import (
	"path/filepath"
	"testing"

	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/site"
)

func TestStarterSiteIsValid(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-docs")
	files, err := Init(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Errorf("files: %v", files)
	}
	s, err := site.Build(site.Options{Root: dir, Dev: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Diags) != 0 {
		for _, d := range s.Diags {
			t.Errorf("starter site: %s", d)
		}
	}
	if s.Config.Title != "My docs" {
		t.Errorf("title = %q", s.Config.Title)
	}
	if _, err := Init(dir); err == nil {
		t.Error("init must not overwrite an existing site")
	}
	if diag.HasErrors(s.Diags) {
		t.Fatal("errors in the starter site")
	}
}
