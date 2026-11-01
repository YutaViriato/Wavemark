// Package diag defines the diagnostics Wavemark reports to content authors.
//
// Authors are frequently AI agents, so every diagnostic carries a stable code,
// a precise location and, when possible, a concrete hint for fixing it.
package diag

import (
	"fmt"
	"sort"
	"strings"
)

// Severity of a diagnostic. Errors block `wavemark serve` and `check`;
// warnings are reported but never block.
type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
)

// Diagnostic codes. They are part of the public contract of `check --json`,
// so existing codes must not be renamed.
const (
	CodeConfigInvalid     = "config.invalid"
	CodeConfigUnknownKey  = "config.unknown-key"
	CodeConfigValue       = "config.value"
	CodeFrontmatter       = "frontmatter.invalid"
	CodeComponentUnknown  = "component.unknown"
	CodeComponentSyntax   = "component.syntax"
	CodeComponentUnclosed = "component.unclosed"
	CodeComponentMismatch = "component.mismatch"
	CodeComponentStray    = "component.stray-close"
	CodeAttrUnknown       = "component.attr-unknown"
	CodeAttrMissing       = "component.attr-missing"
	CodeAttrValue         = "component.attr-value"
	CodeComponentChild    = "component.child"
	CodeComponentParent   = "component.parent"
	CodeComponentLayout   = "component.layout"
	CodeHTMLForbidden     = "html.forbidden"
	CodeTitleDuplicate    = "title.duplicate"
	CodeTitleMultipleH1   = "title.multiple-h1"
	CodeTitleMissing      = "title.missing"
	CodeTitleHidden       = "title.hidden"
	CodeLinkBroken        = "link.broken"
	CodeLinkAnchor        = "link.anchor"
	CodeLinkUnsafe        = "link.unsafe"
	CodeAssetMissing      = "asset.missing"
	CodeRouteConflict     = "route.conflict"
)

// Diagnostic is a single problem found in the content.
type Diagnostic struct {
	File     string   `json:"file"`
	Line     int      `json:"line,omitempty"`   // 1-based; 0 when unknown
	Column   int      `json:"column,omitempty"` // 1-based; 0 when unknown
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Hint     string   `json:"hint,omitempty"`
}

func (d Diagnostic) String() string {
	var b strings.Builder
	b.WriteString(d.File)
	if d.Line > 0 {
		fmt.Fprintf(&b, ":%d", d.Line)
		if d.Column > 0 {
			fmt.Fprintf(&b, ":%d", d.Column)
		}
	}
	fmt.Fprintf(&b, ": %s [%s] %s", d.Severity, d.Code, d.Message)
	if d.Hint != "" {
		b.WriteString("\n    dica: ")
		b.WriteString(d.Hint)
	}
	return b.String()
}

// List collects diagnostics for one file.
type List struct {
	File  string
	Items []Diagnostic
}

// Add appends a diagnostic, filling in the file when it is empty.
func (l *List) Add(d Diagnostic) {
	if d.File == "" {
		d.File = l.File
	}
	if d.Severity == "" {
		d.Severity = Error
	}
	l.Items = append(l.Items, d)
}

// Errorf appends an error at the given position.
func (l *List) Errorf(line, col int, code, hint, format string, args ...any) {
	l.Add(Diagnostic{Line: line, Column: col, Severity: Error, Code: code, Message: fmt.Sprintf(format, args...), Hint: hint})
}

// Warnf appends a warning at the given position.
func (l *List) Warnf(line, col int, code, hint, format string, args ...any) {
	l.Add(Diagnostic{Line: line, Column: col, Severity: Warning, Code: code, Message: fmt.Sprintf(format, args...), Hint: hint})
}

// HasErrors reports whether any diagnostic is an error.
func HasErrors(ds []Diagnostic) bool {
	for _, d := range ds {
		if d.Severity == Error {
			return true
		}
	}
	return false
}

// Count returns the number of errors and warnings.
func Count(ds []Diagnostic) (errors, warnings int) {
	for _, d := range ds {
		if d.Severity == Error {
			errors++
		} else {
			warnings++
		}
	}
	return
}

// Sort orders diagnostics by file, line and column.
func Sort(ds []Diagnostic) {
	sort.SliceStable(ds, func(i, j int) bool {
		a, b := ds[i], ds[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Column < b.Column
	})
}

// Suggest returns the candidate closest to name, or "" when nothing is close
// enough to be a plausible typo.
func Suggest(name string, candidates []string) string {
	best, bestDist := "", 1<<30
	lower := strings.ToLower(name)
	for _, c := range candidates {
		d := levenshtein(lower, strings.ToLower(c))
		if d < bestDist {
			best, bestDist = c, d
		}
	}
	limit := 2
	if len(name) > 8 {
		limit = 3
	}
	if bestDist <= limit {
		return best
	}
	return ""
}

func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}
