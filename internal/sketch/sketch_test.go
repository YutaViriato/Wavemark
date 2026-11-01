package sketch

import (
	"strings"
	"testing"
)

func TestIconsAreDeterministicAndComplete(t *testing.T) {
	for _, name := range Names() {
		a, b := Icon(name, Options{Size: 48}), Icon(name, Options{Size: 48})
		if a == "" || a != b {
			t.Errorf("%s: empty or non-deterministic output", name)
		}
		if !strings.Contains(a, `pathLength="1"`) || strings.Contains(a, "NaN") {
			t.Errorf("%s: invalid path", name)
		}
	}
	for alias, target := range aliases {
		if _, ok := icons[target]; !ok {
			t.Errorf("alias %s points to missing icon %s", alias, target)
		}
	}
	if Icon("rocket-launch", Options{}) != Icon("rocket", Options{}) {
		t.Error("an alias must draw the same icon")
	}
}

func TestLabelMakesIconAccessible(t *testing.T) {
	if s := Icon("book", Options{Label: `Book "x"`}); !strings.Contains(s, `role="img" aria-label="Book &quot;x&quot;"`) {
		t.Errorf("label: %s", s)
	}
	if s := Icon("book", Options{}); !strings.Contains(s, `aria-hidden="true"`) {
		t.Error("an icon without a label must be decorative")
	}
}
