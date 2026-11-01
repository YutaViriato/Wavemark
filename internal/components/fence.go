package components

import (
	"strings"

	"github.com/yuin/goldmark/ast"
)

// FenceAttrs are the attributes accepted in a code fence info string:
//
//	```python title="main.py"
var FenceAttrs = []string{"title"}

// FenceInfo is the parsed info string of a fenced code block.
type FenceInfo struct {
	Lang  string
	Attrs map[string]string
	// Bad lists malformed fragments, for diagnostics.
	Bad []string
}

// ParseFence parses an info string of the form `lang key="value" ...`.
func ParseFence(info string) FenceInfo {
	fi := FenceInfo{Attrs: map[string]string{}}
	s := strings.TrimSpace(info)
	if s == "" {
		return fi
	}
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		fi.Lang, s = s[:i], strings.TrimSpace(s[i:])
	} else {
		fi.Lang, s = s, ""
	}
	if strings.Contains(fi.Lang, "=") {
		s, fi.Lang = strings.TrimSpace(fi.Lang+" "+s), ""
	}
	for s != "" {
		eq := strings.IndexByte(s, '=')
		if eq <= 0 {
			fi.Bad = append(fi.Bad, s)
			break
		}
		key := strings.TrimSpace(s[:eq])
		rest := strings.TrimLeft(s[eq+1:], " \t")
		if rest == "" || (rest[0] != '"' && rest[0] != '\'') {
			fi.Bad = append(fi.Bad, s)
			break
		}
		q := rest[0]
		end := strings.IndexByte(rest[1:], q)
		if end < 0 {
			fi.Bad = append(fi.Bad, s)
			break
		}
		fi.Attrs[key] = rest[1 : end+1]
		s = strings.TrimSpace(rest[end+2:])
	}
	return fi
}

// Fence parses the info string of a fenced code block node.
func Fence(n *ast.FencedCodeBlock, source []byte) FenceInfo {
	if n.Info == nil {
		return FenceInfo{Attrs: map[string]string{}}
	}
	return ParseFence(string(n.Info.Segment.Value(source)))
}

// Label is the text shown in a code block header or CodeGroup tab.
func (fi FenceInfo) Label() string {
	if t := fi.Attrs["title"]; t != "" {
		return t
	}
	return LangName(fi.Lang)
}

var langNames = map[string]string{
	"js": "JavaScript", "javascript": "JavaScript", "ts": "TypeScript", "typescript": "TypeScript",
	"tsx": "TSX", "jsx": "JSX", "py": "Python", "python": "Python", "go": "Go", "rust": "Rust", "rs": "Rust",
	"java": "Java", "kotlin": "Kotlin", "swift": "Swift", "ruby": "Ruby", "rb": "Ruby", "php": "PHP",
	"cs": "C#", "csharp": "C#", "c": "C", "cpp": "C++", "sh": "Shell", "bash": "Bash", "zsh": "Zsh",
	"shell": "Shell", "console": "Terminal", "curl": "cURL", "json": "JSON", "yaml": "YAML", "yml": "YAML",
	"toml": "TOML", "html": "HTML", "css": "CSS", "sql": "SQL", "md": "Markdown", "markdown": "Markdown", "mdx": "MDX",
	"text": "Text", "txt": "Text", "diff": "Diff", "dockerfile": "Dockerfile", "graphql": "GraphQL",
}

// LangName returns a display name for a language identifier.
func LangName(lang string) string {
	if n, ok := langNames[strings.ToLower(lang)]; ok {
		return n
	}
	return lang
}
