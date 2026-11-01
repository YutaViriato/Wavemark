// Package browse builds a navigable snapshot of every Markdown file in a
// project (READMEs, agent instructions, plans, specs, notes), for
// 'wavemark browse'. Unlike a Wavemark site, the files are ordinary
// GitHub-flavoured Markdown: nothing is validated and nothing is rejected.
package browse

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/YutaViriato/Wavemark/internal/config"
	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/i18n"
	"github.com/YutaViriato/Wavemark/internal/markdown"
	"github.com/YutaViriato/Wavemark/internal/site"
	"github.com/YutaViriato/Wavemark/internal/theme"
	"gopkg.in/yaml.v3"
)

// SourcePrefix is the URL prefix of the raw files.
const SourcePrefix = theme.AssetPrefix + "source/"

// maxRender is the largest file rendered; bigger files only link to the source.
const maxRender = 4 << 20

// Options configures Build.
type Options struct {
	Root  string
	Theme *theme.Theme
	// All includes files that .gitignore excludes.
	All bool
	Dev bool
}

// skipDirs are folders that hold dependencies, build output or caches,
// never files someone wrote for the project.
var skipDirs = map[string]bool{
	".git": true, ".hg": true, ".svn": true, "node_modules": true, "bower_components": true, "vendor": true,
	".venv": true, "venv": true, "__pycache__": true, ".tox": true, ".nox": true, ".mypy_cache": true,
	".pytest_cache": true, ".ruff_cache": true, "site-packages": true, ".next": true, ".nuxt": true,
	".svelte-kit": true, ".turbo": true, ".cache": true, ".parcel-cache": true, ".gradle": true,
	".terraform": true, "Pods": true, "target": true, "coverage": true, ".idea": true, ".yarn": true,
	".pnpm-store": true, ".angular": true, ".expo": true, ".dart_tool": true, "DerivedData": true,
}

// SkipDir reports folders that are never scanned or watched.
func SkipDir(_, name string) bool { return skipDirs[name] }

var markdownExts = map[string]bool{".md": true, ".markdown": true, ".mdx": true, ".mdc": true}

var imageExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true, ".avif": true, ".ico": true, ".bmp": true}

// Relevant reports whether a change to rel can change the snapshot.
func Relevant(rel string) bool {
	base := path.Base(rel)
	if strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".swp") || strings.HasPrefix(base, ".#") {
		return false
	}
	ext := strings.ToLower(path.Ext(base))
	return markdownExts[ext] || imageExts[ext] || ext == "" || base == ".gitignore"
}

type file struct {
	rel     string
	abs     string
	size    int64
	modTime time.Time
}

// Build scans the folder and renders every page.
func Build(opts Options) (*site.Site, error) {
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, err
	}
	if fi, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("folder: %w", err)
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("%s is not a folder", root)
	}
	docs, images, err := scan(root, opts.All)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	s := &site.Site{
		Root:    root,
		ByRoute: map[string]*site.Page{},
		ByRaw:   map[string]*site.Page{},
		Assets:  map[string]string{},
		Built:   now,
		T:       i18n.For("en"),
	}
	s.Config.Title = filepath.Base(root)
	s.Config.Lang = "en"
	for _, im := range images {
		s.Assets["/"+im.rel] = im.abs
	}

	b := &builder{site: s, byRel: map[string]*entry{}, dirs: map[string]bool{".": true}, now: now}
	for _, f := range docs {
		b.dirs[path.Dir(f.rel)] = true
		for d := path.Dir(f.rel); d != "."; d = path.Dir(d) {
			b.dirs[d] = true
		}
	}
	entries := make([]*entry, len(docs))
	for i, f := range docs {
		entries[i] = b.load(f)
		b.byRel[f.rel] = entries[i]
	}
	for _, e := range entries {
		s.ByRoute[e.page.Route] = e.page
		s.ByRaw[e.view.SourceURL] = e.page
	}
	s.Nav = b.nav(entries)
	for _, e := range entries {
		e.page.Body = b.render(e)
		s.Pages = append(s.Pages, e.page)
	}
	if err := b.renderAll(opts, entries); err != nil {
		return nil, err
	}
	s.Version = version(docs, images)
	return s, nil
}

type entry struct {
	page *site.Page
	view *theme.FileView
	doc  *markdown.Document
	file file
}

type builder struct {
	site  *site.Site
	byRel map[string]*entry
	dirs  map[string]bool
	now   time.Time
}

func (b *builder) load(f file) *entry {
	p := &site.Page{Rel: f.rel, Abs: f.abs, Route: "/" + f.rel}
	p.Settings = config.Resolve(config.Defaults{}, config.Frontmatter{})
	v := &theme.FileView{Path: f.rel, Name: path.Base(f.rel), Route: p.Route, SourceURL: SourcePrefix + f.rel, Modified: f.modTime, Badges: badges(f.rel)}
	e := &entry{page: p, view: v, file: f}
	p.Crumbs = b.crumbs(path.Dir(f.rel))

	src, err := os.ReadFile(f.abs)
	if err != nil {
		src = []byte(fmt.Sprintf("Cannot read this file: %v\n", err))
	}
	p.Source = src
	if len(src) > maxRender {
		v.TooLarge = true
		p.Title, v.Title = v.Name, v.Name
		v.Words = len(bytes.Fields(src))
		return e
	}
	if !utf8.Valid(src) {
		src = bytes.ToValidUTF8(src, []byte("\uFFFD"))
	}
	fm, body, _, bodyOffset := config.SplitFrontmatter(src, &diag.List{})
	facts, fmTitle, fmDesc := frontmatter(fm)
	v.Facts = facts
	e.doc = markdown.ParsePlain(body, bodyOffset)
	a := markdown.Analyze(e.doc, "docs", fmTitle != "", &diag.List{})
	p.Headings, p.IDs = a.Headings, a.IDs
	switch {
	case fmTitle != "":
		p.Title = fmTitle
	case a.Title != "":
		p.Title = a.Title
	default:
		p.Title = v.Name
	}
	p.Description = fmDesc
	p.Summary = fmDesc
	if p.Summary == "" {
		p.Summary = a.Plain
	}
	v.Title, v.Summary = p.Title, p.Summary
	v.Words = len(bytes.Fields(body))
	return e
}

// frontmatter returns the top-level keys of a YAML frontmatter as text,
// plus its title and description. Invalid YAML is shown as one fact.
func frontmatter(data []byte) (facts []theme.Fact, title, desc string) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, "", ""
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return []theme.Fact{{Key: "frontmatter", Value: oneLine(string(data), 240)}}, "", ""
	}
	m := doc.Content[0]
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i].Value, m.Content[i+1]
		text := ""
		switch v.Kind {
		case yaml.ScalarNode:
			text = v.Value
		case yaml.SequenceNode:
			var items []string
			for _, it := range v.Content {
				if it.Kind == yaml.ScalarNode {
					items = append(items, it.Value)
				} else {
					out, _ := yaml.Marshal(it)
					items = append(items, oneLine(string(out), 80))
				}
			}
			text = strings.Join(items, ", ")
		default:
			out, _ := yaml.Marshal(v)
			text = oneLine(string(out), 240)
		}
		switch strings.ToLower(k) {
		case "title":
			if v.Kind == yaml.ScalarNode {
				title = strings.TrimSpace(v.Value)
			}
		case "description":
			if v.Kind == yaml.ScalarNode {
				desc = strings.TrimSpace(v.Value)
			}
		}
		switch strings.ToLower(k) {
		case "title", "description":
			if v.Kind == yaml.ScalarNode {
				continue // shown as the page title and lead
			}
		}
		facts = append(facts, theme.Fact{Key: k, Value: text})
	}
	return facts, title, desc
}

func oneLine(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > max {
		s = string(r[:max]) + "…"
	}
	return s
}

// crumbs lists the folders above a file or folder, each linking to its
// listing.
func (b *builder) crumbs(dir string) []site.Crumb {
	if dir == "." || dir == "" {
		return []site.Crumb{{Title: b.site.Config.Title, Route: "/"}}
	}
	out := b.crumbs(path.Dir(dir))
	return append(out, site.Crumb{Title: path.Base(dir), Route: "/" + dir})
}

// instructionFiles are the conventional names of agent instruction files.
// They are matched case-sensitively: site/agents.md is a page about agents,
// AGENTS.md is instructions for them.
var instructionFiles = map[string]bool{"CLAUDE.md": true, "CLAUDE.local.md": true, "AGENTS.md": true, "AGENT.md": true, "GEMINI.md": true}

// badges marks files that steer or describe agent work.
func badges(rel string) []theme.Badge {
	lower := strings.ToLower(rel)
	base := path.Base(lower)
	in := func(dirs ...string) bool {
		for _, d := range dirs {
			if strings.HasPrefix(lower, d) || strings.Contains(lower, "/"+d) {
				return true
			}
		}
		return false
	}
	switch {
	case base == "skill.md" || in(".claude/skills/", ".kiro/skills/", ".agents/skills/"):
		return []theme.Badge{{Label: "Skill", Tone: "olive"}}
	case in(".claude/commands/", ".kiro/prompts/", ".github/prompts/"):
		return []theme.Badge{{Label: "Command", Tone: "fig"}}
	case in(".claude/agents/", ".kiro/agents/", ".github/agents/"):
		return []theme.Badge{{Label: "Agent", Tone: "fig"}}
	case in(".kiro/specs/"):
		return []theme.Badge{{Label: "Spec", Tone: "sky"}}
	case instructionFiles[path.Base(rel)] || base == "copilot-instructions.md" || strings.HasSuffix(base, ".instructions.md") ||
		in(".kiro/steering/", ".cursor/rules/", ".windsurf/rules/", ".clinerules/", ".github/instructions/", ".amazonq/rules/"):
		return []theme.Badge{{Label: "Instructions", Tone: "clay"}}
	}
	return nil
}

// render renders a page body. Relative links that point at another file of
// the project become absolute routes; anything else is left as written.
func (b *builder) render(e *entry) string {
	if e.doc == nil {
		return ""
	}
	dir := path.Dir(e.page.Rel)
	url := func(dest string) string { return b.resolve(dir, dest) }
	body, err := markdown.Render(e.doc, markdown.RenderOptions{URL: url, T: b.site.T, Layout: "docs", Lang: "en", Plain: true})
	if err != nil {
		return "<p>Cannot render this file: " + escape(err.Error()) + "</p>"
	}
	e.doc = nil
	return body
}

func (b *builder) resolve(dir, dest string) string {
	if dest == "" || strings.HasPrefix(dest, "#") || strings.HasPrefix(dest, "/") || strings.HasPrefix(dest, "//") || strings.Contains(strings.SplitN(dest, "/", 2)[0], ":") {
		return dest
	}
	target, frag := dest, ""
	if i := strings.IndexAny(target, "#?"); i >= 0 {
		target, frag = target[:i], target[i:]
	}
	clean := path.Join(dir, target)
	if strings.HasPrefix(clean, "../") || clean == ".." {
		return dest // outside the browsed folder
	}
	if _, ok := b.byRel[clean]; ok {
		return "/" + clean + frag
	}
	if _, ok := b.site.Assets["/"+clean]; ok {
		return "/" + clean + frag
	}
	if b.dirs[clean] {
		if clean == "." {
			return "/" + frag
		}
		return "/" + clean + frag
	}
	return dest
}

func escape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

// nav builds the sidebar: files first (README first), then folders.
func (b *builder) nav(entries []*entry) *site.NavNode {
	root := &site.NavNode{Dir: ".", Title: b.site.Config.Title}
	nodes := map[string]*site.NavNode{".": root}
	var ensure func(dir string) *site.NavNode
	ensure = func(dir string) *site.NavNode {
		if n, ok := nodes[dir]; ok {
			return n
		}
		parent := ensure(path.Dir(dir))
		n := &site.NavNode{Dir: dir, Title: path.Base(dir)}
		parent.Children = append(parent.Children, n)
		nodes[dir] = n
		return n
	}
	for _, e := range entries {
		n := ensure(path.Dir(e.page.Rel))
		n.Children = append(n.Children, &site.NavNode{Title: e.view.Name, Page: e.page})
	}
	var sortNode func(n *site.NavNode)
	sortNode = func(n *site.NavNode) {
		sort.SliceStable(n.Children, func(i, j int) bool { return navLess(n.Children[i], n.Children[j]) })
		for _, c := range n.Children {
			if c.IsSection() {
				sortNode(c)
			}
		}
	}
	sortNode(root)
	return root
}

func navLess(a, b *site.NavNode) bool {
	if a.IsSection() != b.IsSection() {
		return !a.IsSection()
	}
	ra, rb := rank(a.Title), rank(b.Title)
	if ra != rb {
		return ra < rb
	}
	return strings.ToLower(a.Title) < strings.ToLower(b.Title)
}

// rank puts the files people open first at the top of their folder.
func rank(name string) int {
	switch strings.ToLower(strings.TrimSuffix(name, path.Ext(name))) {
	case "readme", "index":
		return 0
	case "agents", "claude", "gemini":
		return 1
	}
	return 2
}

// renderAll renders the file pages and one listing per folder.
func (b *builder) renderAll(opts Options, entries []*entry) error {
	th := opts.Theme
	topts := theme.Options{Dev: opts.Dev}
	for _, e := range entries {
		html, err := th.RenderFile(b.site, e.page, e.view, b.now, topts)
		if err != nil {
			return fmt.Errorf("rendering %s: %w", e.page.Rel, err)
		}
		setHTML(e.page, html)
	}
	dirs := make([]string, 0, len(b.dirs))
	for d := range b.dirs {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		l := b.listing(dir, entries)
		route := "/"
		title := b.site.Config.Title
		if dir != "." {
			route, title = "/"+dir, path.Base(dir)
		}
		if _, taken := b.site.ByRoute[route]; taken {
			continue
		}
		p := &site.Page{Rel: dir, Route: route, Title: title, Settings: config.Resolve(config.Defaults{}, config.Frontmatter{})}
		if dir != "." {
			p.Crumbs = b.crumbs(path.Dir(dir))
		}
		html, err := th.RenderListing(b.site, p, l, topts)
		if err != nil {
			return fmt.Errorf("rendering the listing of %s: %w", dir, err)
		}
		setHTML(p, html)
		b.site.ByRoute[route] = p
	}
	return nil
}

func setHTML(p *site.Page, html []byte) {
	p.HTML = html
	sum := sha256.Sum256(html)
	p.ETag = `"` + hex.EncodeToString(sum[:8]) + `"`
}

// listing collects the files under dir, grouped by folder.
func (b *builder) listing(dir string, entries []*entry) *theme.ListingView {
	l := &theme.ListingView{Root: b.site.Root, Now: b.now, Title: b.site.Config.Title}
	if dir != "." {
		l.Path, l.Title = dir, path.Base(dir)
		l.Root = filepath.Join(b.site.Root, filepath.FromSlash(dir))
	}
	groups := map[string]*theme.FolderView{}
	var files []*theme.FileView
	for _, e := range entries {
		fdir := path.Dir(e.page.Rel)
		if dir != "." && fdir != dir && !strings.HasPrefix(fdir, dir+"/") {
			continue
		}
		g, ok := groups[fdir]
		if !ok {
			g = &theme.FolderView{Route: "/"}
			if fdir != "." {
				g.Path, g.Route = fdir, "/"+fdir
			}
			groups[fdir] = g
		}
		g.Files = append(g.Files, e.view)
		files = append(files, e.view)
		if e.view.Modified.After(l.Updated) {
			l.Updated = e.view.Modified
		}
	}
	l.Files, l.Folders = len(files), len(groups)
	for _, g := range groups {
		sort.SliceStable(g.Files, func(i, j int) bool {
			ri, rj := rank(g.Files[i].Name), rank(g.Files[j].Name)
			if ri != rj {
				return ri < rj
			}
			return strings.ToLower(g.Files[i].Name) < strings.ToLower(g.Files[j].Name)
		})
		l.Groups = append(l.Groups, g)
	}
	sort.Slice(l.Groups, func(i, j int) bool { return l.Groups[i].Path < l.Groups[j].Path })
	if len(files) > 4 {
		recent := append([]*theme.FileView(nil), files...)
		sort.SliceStable(recent, func(i, j int) bool { return recent[i].Modified.After(recent[j].Modified) })
		l.Recent = recent[:min(6, len(recent))]
	}
	return l
}

// version fingerprints what the pages show, so saving an unrelated file
// does not reload open pages.
func version(groups ...[]file) string {
	h := sha256.New()
	for _, fs := range groups {
		for _, f := range fs {
			fmt.Fprintf(h, "%s\x00%d\x00%d\n", f.rel, f.size, f.modTime.UnixNano())
		}
	}
	return hex.EncodeToString(h.Sum(nil)[:12])
}

// scan finds the Markdown files and images. Inside a git repository it asks
// git, so .gitignore applies (unless all is set); otherwise, or when git
// finds no Markdown at all (the folder itself is ignored), it walks the
// folder. Dependency and cache folders are always skipped and symlinks are
// never followed.
func scan(root string, all bool) (docs, images []file, err error) {
	if !all {
		if rels, err := gitFiles(root); err == nil {
			docs, images = collect(root, rels)
			if len(docs) > 0 {
				return docs, images, nil
			}
		}
	}
	rels, err := walk(root)
	if err != nil {
		return nil, nil, err
	}
	docs, images = collect(root, rels)
	return docs, images, nil
}

func collect(root string, rels []string) (docs, images []file) {
	seen := map[string]bool{}
	for _, rel := range rels {
		if seen[rel] || skipped(rel) {
			continue
		}
		seen[rel] = true
		ext := strings.ToLower(path.Ext(rel))
		if !markdownExts[ext] && !imageExts[ext] {
			continue
		}
		abs := filepath.Join(root, filepath.FromSlash(rel))
		fi, err := os.Lstat(abs)
		if err != nil || !fi.Mode().IsRegular() {
			continue // deleted, a symlink or not a file
		}
		f := file{rel: rel, abs: abs, size: fi.Size(), modTime: fi.ModTime()}
		if markdownExts[ext] {
			docs = append(docs, f)
		} else {
			images = append(images, f)
		}
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].rel < docs[j].rel })
	sort.Slice(images, func(i, j int) bool { return images[i].rel < images[j].rel })
	return docs, images
}

func skipped(rel string) bool {
	parts := strings.Split(rel, "/")
	for _, p := range parts[:len(parts)-1] {
		if skipDirs[p] {
			return true
		}
	}
	return false
}

// gitFiles lists tracked and untracked, not ignored files under root.
func gitFiles(root string) ([]string, error) {
	cmd := exec.Command("git", "-C", root, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var rels []string
	for _, p := range bytes.Split(out, []byte{0}) {
		if len(p) > 0 {
			rels = append(rels, string(p))
		}
	}
	return rels, nil
}

func walk(root string) ([]string, error) {
	var rels []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == root {
				return err
			}
			return nil // unreadable folders are skipped, not fatal
		}
		if d.IsDir() {
			if p != root && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 || !d.Type().IsRegular() {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		rels = append(rels, filepath.ToSlash(rel))
		return nil
	})
	return rels, err
}
