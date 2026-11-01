package site

import (
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/config"
	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/markdown"
)

// resolution is the outcome of resolving a destination.
type resolution struct {
	href       string
	code, msg  string
	hint       string
	isExternal bool
}

func (r resolution) ok() bool { return r.code == "" }

// resolve maps a destination as written in a page (or in wavemark.yaml when
// from is the root pseudo-page) to its final href.
func (s *Site) resolve(from *Page, dest string, asset bool) resolution {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return resolution{code: diag.CodeLinkBroken, msg: "empty destination", hint: "Fill in the link destination."}
	}
	if strings.HasPrefix(dest, "//") {
		return resolution{href: "https:" + dest, isExternal: true}
	}
	if u, err := url.Parse(dest); err == nil && u.Scheme != "" {
		switch strings.ToLower(u.Scheme) {
		case "http", "https":
			return resolution{href: dest, isExternal: true}
		case "mailto", "tel":
			if !asset {
				return resolution{href: dest, isExternal: true}
			}
		}
		return resolution{code: diag.CodeLinkUnsafe, msg: fmt.Sprintf("URL scheme not allowed: %s", u.Scheme), hint: "Use http(s), mailto, tel, a route (/guide) or a relative file (./guide.md)."}
	}

	p, frag, _ := strings.Cut(dest, "#")
	p, query, _ := strings.Cut(p, "?")
	if dec, err := url.PathUnescape(p); err == nil {
		p = dec
	}
	if dec, err := url.PathUnescape(frag); err == nil {
		frag = dec
	}

	if p == "" { // same-page anchor
		if asset {
			return resolution{code: diag.CodeAssetMissing, msg: "empty file destination", hint: "Point to a file, such as ./image.png."}
		}
		if from != nil && from.IDs != nil && !from.IDs[frag] {
			return anchorProblem(from, frag)
		}
		return resolution{href: "#" + frag}
	}

	var clean string
	if strings.HasPrefix(p, "/") {
		clean = strings.TrimPrefix(path.Clean(p), "/")
	} else {
		dir := "."
		if from != nil {
			dir = from.dir
		}
		clean = path.Join(dir, p)
	}
	if clean == "." {
		clean = ""
	}
	if strings.HasPrefix(clean, "../") || clean == ".." {
		return resolution{code: diag.CodeLinkBroken, msg: fmt.Sprintf("%s points outside the site folder", dest), hint: "Links can only point to files inside the site."}
	}
	suffix := ""
	if query != "" {
		suffix = "?" + query
	}

	if asset {
		if _, ok := s.Assets["/"+clean]; ok {
			return resolution{href: "/" + clean + suffix}
		}
		return resolution{code: diag.CodeAssetMissing, msg: fmt.Sprintf("file %s does not exist", "/"+clean), hint: s.assetHint(clean)}
	}

	target := s.findPage(clean, strings.HasPrefix(p, "/"))
	if target == nil {
		if clean == "llms.txt" {
			return resolution{href: "/llms.txt"} // served by the runtime
		}
		if _, ok := s.Assets["/"+clean]; ok {
			return resolution{href: "/" + clean + suffix}
		}
		return resolution{code: diag.CodeLinkBroken, msg: fmt.Sprintf("link to a page that does not exist: %s", dest), hint: s.pageHint(clean)}
	}
	href := target.Route + suffix
	if frag != "" {
		if !target.IDs[frag] {
			return anchorProblem(target, frag)
		}
		href += "#" + frag
	}
	return resolution{href: href}
}

// findPage looks a cleaned path up as a file path first, then as a route.
func (s *Site) findPage(clean string, absolute bool) *Page {
	if strings.EqualFold(path.Ext(clean), ".md") {
		return s.byRel[clean]
	}
	for _, c := range []string{clean + ".md", path.Join(clean, "index.md"), path.Join(clean, "README.md")} {
		if p := s.byRel[strings.TrimPrefix(c, "./")]; p != nil {
			return p
		}
	}
	r := "/" + strings.TrimSuffix(clean, "/")
	if r == "/" || clean == "" {
		r = "/"
	}
	return s.ByRoute[r]
}

func anchorProblem(target *Page, frag string) resolution {
	ids := make([]string, 0, len(target.IDs))
	for id := range target.IDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	hint := "Anchors come from headings (## My Section → #my-section)."
	if len(ids) > 0 {
		if sug := diag.Suggest(frag, ids); sug != "" {
			hint = fmt.Sprintf("Did you mean #%s? %s", sug, hint)
		} else {
			hint += " Existing anchors: #" + strings.Join(ids, ", #") + "."
		}
	}
	return resolution{code: diag.CodeLinkAnchor, msg: fmt.Sprintf("anchor #%s does not exist in %s", frag, target.Rel), hint: hint}
}

func (s *Site) pageHint(clean string) string {
	var cands []string
	for rel := range s.byRel {
		cands = append(cands, rel)
	}
	want := clean
	if !strings.EqualFold(path.Ext(want), ".md") {
		want += ".md"
	}
	if sug := diag.Suggest(want, cands); sug != "" {
		return fmt.Sprintf("Did you mean %s? Relative links start from the page's folder; use the path of the .md file.", sug)
	}
	return "Relative links start from the page's folder (./other.md, ../guide/x.md). Absolute links use the route (/guide/x)."
}

func (s *Site) assetHint(clean string) string {
	var cands []string
	for a := range s.Assets {
		cands = append(cands, strings.TrimPrefix(a, "/"))
	}
	if sug := diag.Suggest(clean, cands); sug != "" {
		return fmt.Sprintf("Did you mean %s? Relative paths start from the page's folder.", sug)
	}
	return "Put the file inside the site folder. Relative paths start from the page's folder."
}

// checkRefs validates every link and asset reference of a page.
func (s *Site) checkRefs(p *Page) {
	if p.analysis == nil {
		return
	}
	for _, ref := range p.analysis.Refs {
		if ref.Type == markdown.RefDir {
			if _, ok := s.resolveDir(p, ref.Dest); !ok {
				addDiag(p, diag.Diagnostic{Line: ref.Line, Column: ref.Col, Code: diag.CodeLinkBroken, Message: ref.Where + ": folder " + ref.Dest + " does not exist", Hint: "Use a folder path relative to the page, such as ./blog, or absolute, such as /blog."})
			}
			continue
		}
		r := s.resolve(p, ref.Dest, ref.Type == markdown.RefAsset)
		if r.ok() {
			continue
		}
		addDiag(p, diag.Diagnostic{Line: ref.Line, Column: ref.Col, Code: r.code, Message: ref.Where + ": " + r.msg, Hint: r.hint})
	}
}

// urlFor returns the destination mapper used while rendering p.
func (s *Site) urlFor(p *Page) func(string) string {
	return func(dest string) string {
		asset := looksLikeAsset(dest)
		r := s.resolve(p, dest, false)
		if !r.ok() && asset {
			r = s.resolve(p, dest, true)
		}
		if r.code == diag.CodeLinkUnsafe {
			return "" // never emit javascript:, data: and friends, even in dev
		}
		if !r.ok() {
			return dest
		}
		return r.href
	}
}

func looksLikeAsset(dest string) bool {
	ext := strings.ToLower(path.Ext(strings.SplitN(dest, "#", 2)[0]))
	return ext != "" && ext != ".md"
}

var rootPage = &Page{dir: "."}

func (s *Site) resolveLinks(links []config.NavLink, key string, dl *diag.List) []NavLink {
	var out []NavLink
	for i, n := range links {
		if n.Href == "" {
			continue // reported by config
		}
		r := s.resolve(rootPage, n.Href, false)
		if !r.ok() {
			dl.Add(diag.Diagnostic{Code: r.code, Message: fmt.Sprintf("%s[%d] (%s): %s", key, i, n.Label, r.msg), Hint: r.hint})
			continue
		}
		out = append(out, NavLink{Label: n.Label, Href: r.href})
	}
	return out
}

func (s *Site) resolveFonts(dl *diag.List) Fonts {
	var f Fonts
	one := func(key, p string) string {
		if p == "" {
			return ""
		}
		if !strings.EqualFold(path.Ext(p), ".woff2") {
			dl.Errorf(0, 0, diag.CodeConfigValue, "Convert the font to woff2.", "fonts.%s: only .woff2 files are accepted", key)
			return ""
		}
		r := s.resolve(rootPage, p, true)
		if !r.ok() || r.isExternal {
			dl.Errorf(0, 0, diag.CodeAssetMissing, "Put the file inside the site folder and use a path relative to the root.", "fonts.%s: %s", key, r.msg)
			return ""
		}
		return r.href
	}
	f.Sans = one("sans", s.Config.Fonts.Sans)
	f.Serif = one("serif", s.Config.Fonts.Serif)
	f.Mono = one("mono", s.Config.Fonts.Mono)
	if src := s.Config.Logo.Src; src != "" {
		if r := s.resolve(rootPage, src, true); !r.ok() {
			dl.Errorf(0, 0, r.code, r.hint, "logo.src: %s", r.msg)
		}
	}
	return f
}

// LogoSrc resolves the logo image, or "" when none is configured or found.
func (s *Site) LogoSrc() string {
	if s.Config.Logo.Src == "" {
		return ""
	}
	r := s.resolve(rootPage, s.Config.Logo.Src, true)
	if !r.ok() {
		return ""
	}
	return r.href
}
