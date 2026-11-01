---
title: Markdown
description: The Markdown Wavemark accepts and how each element looks in the theme.
---

Wavemark accepts CommonMark with the GitHub extensions (GFM): tables, strikethrough, task lists, autolinks and footnotes.

## Text

Paragraphs use the 15 px sans body. **Bold** takes the strong text color, *italic* keeps the body color and ~~strikethrough~~ is muted. Inline code like `wavemark check` uses the mono font. Internal links like [the installation guide](../guide/install.md) and external ones like [CommonMark](https://commonmark.org) have a quiet underline.[^1]

> Blockquotes are indented, with a border on the left.

## Lists

- List items
- With muted markers
  - And nested levels

1. Numbered lists
2. Follow the same rhythm

- [x] A finished task
- [ ] A pending task

## Tables

| Command | What it is for | Fails on errors? |
|:---|:---|:---:|
| `wavemark dev` | Writing, with live reload | No |
| `wavemark check` | Validating before publishing | Yes |
| `wavemark serve` | Production | Refuses to start |

## Code

```go title="internal/site/site.go"
// Build scans the folder and produces a snapshot.
func Build(opts Options) (*Site, error) {
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, err
	}
	return &Site{Root: root}, nil
}
```

```json
{
  "ok": false,
  "errors": 1,
  "diagnostics": [{ "file": "guide/x.md", "line": 12, "code": "link.broken" }]
}
```

```
A block without a language: plain text, no highlighting.
```

### Third-level headings

`###` headings are in the table of contents too, indented.

#### Fourth level

`####` headings use bold sans and stay out of the table of contents.

---

[^1]: Footnotes appear at the end of the page.
