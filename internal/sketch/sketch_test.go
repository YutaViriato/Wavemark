package sketch

import (
	"strings"
	"testing"
)

func TestIconsAreDeterministicAndComplete(t *testing.T) {
	for _, name := range Names() {
		a, b := Icon(name, Options{Size: 48}), Icon(name, Options{Size: 48})
		if a == "" || a != b {
			t.Errorf("%s: saída vazia ou não determinística", name)
		}
		if !strings.Contains(a, `pathLength="1"`) || strings.Contains(a, "NaN") {
			t.Errorf("%s: path inválido", name)
		}
	}
	for alias, target := range aliases {
		if _, ok := icons[target]; !ok {
			t.Errorf("alias %s aponta para ícone inexistente %s", alias, target)
		}
	}
	if Icon("rocket-launch", Options{}) != Icon("rocket", Options{}) {
		t.Error("alias deve desenhar o mesmo ícone")
	}
}

func TestLabelMakesIconAccessible(t *testing.T) {
	if s := Icon("book", Options{Label: `Livro "x"`}); !strings.Contains(s, `role="img" aria-label="Livro &quot;x&quot;"`) {
		t.Errorf("rótulo: %s", s)
	}
	if s := Icon("book", Options{}); !strings.Contains(s, `aria-hidden="true"`) {
		t.Error("ícone sem rótulo deve ser decorativo")
	}
}
