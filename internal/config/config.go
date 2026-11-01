// Package config loads and validates the three configuration levels of a
// Wavemark site: wavemark.yaml (site), _meta.yaml (folder) and the YAML
// frontmatter of each page.
//
// Cascade: site.defaults → folder.defaults (outermost to innermost) → page.
// A value set at a deeper level overrides the shallower one. Unknown keys
// are errors, so typos never fail silently.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/YutaViriato/Wavemark/internal/diag"
	"gopkg.in/yaml.v3"
)

// File names with special meaning.
const (
	SiteFile   = "wavemark.yaml"
	FolderFile = "_meta.yaml"
)

// Layouts known to the theme.
var Layouts = []string{"docs", "landing", "article"}

// Defaults holds the page settings that cascade from site to folder to page.
// Pointers distinguish "not set" from the zero value.
type Defaults struct {
	Layout *string `yaml:"layout" doc:"Layout da página: docs (padrão; navegação lateral), article (leitura longa, como os artigos do anthropic.com) ou landing (página de entrada)."`
	TOC    *bool   `yaml:"toc" doc:"Mostra o sumário (TOC) à direita no layout docs. Padrão: true."`
	Hidden *bool   `yaml:"hidden" doc:"Esconde a página (ou a pasta inteira, se usado em defaults) da navegação. A página continua acessível pela URL."`
}

// Merge returns d overridden by every field set in over.
func (d Defaults) Merge(over Defaults) Defaults {
	if over.Layout != nil {
		d.Layout = over.Layout
	}
	if over.TOC != nil {
		d.TOC = over.TOC
	}
	if over.Hidden != nil {
		d.Hidden = over.Hidden
	}
	return d
}

// Frontmatter is the YAML block at the top of a page.
type Frontmatter struct {
	Title       *string `yaml:"title" doc:"Título da página. Se omitido, o primeiro '# H1' do corpo vira o título."`
	Description *string `yaml:"description" doc:"Resumo exibido abaixo do título e usado na meta description e no llms.txt."`
	Order       *int    `yaml:"order" doc:"Posição na navegação dentro da pasta (menor vem primeiro). O 'order' do _meta.yaml tem precedência."`
	Date        *string `yaml:"date" doc:"Data de publicação no formato AAAA-MM-DD. Aparece no cabeçalho e permite 'sort: date' no _meta.yaml."`
	Author      *string `yaml:"author" doc:"Autor ou autores, como texto (ex.: \"Ana Lima e Rui Costa\")."`
	Icon        *string `yaml:"icon" doc:"Ícone desenhado à mão da página, usado por <PageList style=\"cards\">."`
	Defaults    `yaml:",inline"`
}

// Folder is the content of a _meta.yaml file.
type Folder struct {
	Title    string   `yaml:"title" doc:"Nome da seção na navegação. Padrão: nome da pasta."`
	Order    []string `yaml:"order" doc:"Ordem explícita dos itens da pasta: nomes de arquivo sem .md ou nomes de subpasta. Itens não listados vêm depois."`
	Sort     string   `yaml:"sort" doc:"Ordem dos itens não listados em 'order': title (padrão) ou date (mais recentes primeiro, pela chave 'date' das páginas)."`
	Defaults Defaults `yaml:"defaults" doc:"Configurações herdadas por todas as páginas da pasta e das subpastas."`
}

// Logo configures the brand mark in the header.
type Logo struct {
	Text string `yaml:"text" doc:"Texto do logo (padrão: title do site)."`
	Src  string `yaml:"src" doc:"Imagem do logo, relativa à raiz do site (SVG ou PNG)."`
}

// UnmarshalYAML accepts `logo: ./img/logo.svg` as a shorthand for
// `logo: {src: ./img/logo.svg}`, which agents write naturally.
func (l *Logo) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		l.Src = n.Value
		return nil
	}
	if n.Kind != yaml.MappingNode {
		return &yaml.TypeError{Errors: []string{fmt.Sprintf("line %d: cannot unmarshal !!seq `logo` into config.Logo", n.Line)}}
	}
	var errs []string
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		switch k.Value {
		case "text":
			l.Text = v.Value
		case "src":
			l.Src = v.Value
		default:
			errs = append(errs, fmt.Sprintf("line %d: field %s not found in type config.Logo", k.Line, k.Value))
		}
	}
	if len(errs) > 0 {
		return &yaml.TypeError{Errors: errs}
	}
	return nil
}

// NavLink is an entry of the top navigation.
type NavLink struct {
	Label string `yaml:"label" doc:"Texto do link."`
	Href  string `yaml:"href" doc:"Destino: rota (/guia), arquivo .md relativo à raiz ou URL https."`
}

// Footer configures the site footer.
type Footer struct {
	Text  string    `yaml:"text" doc:"Texto do rodapé, por exemplo o copyright."`
	Links []NavLink `yaml:"links" doc:"Links do rodapé, com 'label' e 'href'."`
}

// Fonts lets a site use font files it is licensed for.
type Fonts struct {
	Sans  string `yaml:"sans" doc:"Arquivo woff2 (relativo à raiz) para textos sans-serif."`
	Serif string `yaml:"serif" doc:"Arquivo woff2 (relativo à raiz) para títulos serifados."`
	Mono  string `yaml:"mono" doc:"Arquivo woff2 (relativo à raiz) para código."`
}

// Site is the content of wavemark.yaml.
type Site struct {
	Title       string    `yaml:"title" doc:"Nome do site, usado no header e no <title>."`
	Description string    `yaml:"description" doc:"Descrição do site para meta tags e llms.txt."`
	Lang        string    `yaml:"lang" doc:"Idioma do conteúdo (atributo lang do HTML). Padrão: en."`
	BaseURL     string    `yaml:"baseURL" doc:"URL pública do site (https://...), usada para links absolutos no llms.txt."`
	Logo        Logo      `yaml:"logo" doc:"Logo do header. Mapa com 'text' (texto ao lado do logo; padrão: title) e 'src' (imagem relativa à raiz), ou só o caminho da imagem: logo: ./img/logo.svg."`
	Nav         []NavLink `yaml:"nav" doc:"Links da navegação superior. Cada item tem 'label' e 'href'."`
	Footer      Footer    `yaml:"footer" doc:"Rodapé: 'text' (texto simples, sem Markdown) e 'links' (lista de itens com 'label' e 'href')."`
	Fonts       Fonts     `yaml:"fonts" doc:"Fontes próprias: 'sans', 'serif' e 'mono', cada uma um arquivo .woff2 relativo à raiz."`
	Defaults    Defaults  `yaml:"defaults" doc:"Configurações herdadas por todas as páginas."`
}

// Settings are the effective settings of a page after the cascade.
type Settings struct {
	Title       string
	Description string
	Date        time.Time // zero when absent
	Author      string
	Icon        string
	Layout      string
	TOC         bool
	Hidden      bool
	Order       int
	HasOrder    bool
}

// Resolve applies the cascade and fills defaults.
func Resolve(cascade Defaults, fm Frontmatter) Settings {
	eff := cascade.Merge(fm.Defaults)
	s := Settings{Layout: "docs", TOC: true}
	if eff.Layout != nil {
		s.Layout = *eff.Layout
	}
	if eff.TOC != nil {
		s.TOC = *eff.TOC
	}
	if eff.Hidden != nil {
		s.Hidden = *eff.Hidden
	}
	if fm.Title != nil {
		s.Title = strings.TrimSpace(*fm.Title)
	}
	if fm.Description != nil {
		s.Description = strings.TrimSpace(*fm.Description)
	}
	if fm.Order != nil {
		s.Order, s.HasOrder = *fm.Order, true
	}
	if fm.Author != nil {
		s.Author = strings.TrimSpace(*fm.Author)
	}
	if fm.Icon != nil {
		s.Icon = strings.TrimSpace(*fm.Icon)
	}
	if fm.Date != nil {
		s.Date, _ = ParseDate(*fm.Date)
	}
	return s
}

// ParseDate accepts YYYY-MM-DD (optionally followed by a time).
func ParseDate(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02 15:04", "2006-01-02T15:04"} {
		if t, err := time.Parse(layout, v); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("data inválida %q", v)
}

// LoadSite parses wavemark.yaml. A missing file yields an empty config.
func LoadSite(data []byte, diags *diag.List) Site {
	var s Site
	root := decodeStrict(data, 0, &s, diags)
	validateDefaults(root, []string{"defaults"}, s.Defaults, 0, diags)
	unknownAt := map[int]bool{}
	for _, d := range diags.Items {
		if d.Code == diag.CodeConfigUnknownKey {
			unknownAt[d.Line] = true
		}
	}
	for i, n := range s.Nav {
		line := keyLine(root, 0, "nav", strconv.Itoa(i))
		next := keyLine(root, 0, "nav", strconv.Itoa(i+1))
		if i+1 >= len(s.Nav) {
			next = line + 1000
		}
		// A misspelled key ('title' instead of 'label') was already reported.
		misspelled := false
		for l := line; l < next; l++ {
			misspelled = misspelled || unknownAt[l]
		}
		if strings.TrimSpace(n.Label) == "" && !misspelled {
			diags.Errorf(line, 0, diag.CodeConfigValue, "", "nav[%d]: 'label' é obrigatório", i)
		}
		if strings.TrimSpace(n.Href) == "" && !misspelled {
			diags.Errorf(line, 0, diag.CodeConfigValue, "", "nav[%d]: 'href' é obrigatório", i)
		}
	}
	if s.BaseURL != "" && !strings.HasPrefix(s.BaseURL, "https://") && !strings.HasPrefix(s.BaseURL, "http://") {
		diags.Errorf(keyLine(root, 0, "baseURL"), 0, diag.CodeConfigValue, "Exemplo: baseURL: https://docs.exemplo.com", "baseURL precisa ser uma URL http(s) absoluta")
	}
	s.BaseURL = strings.TrimRight(s.BaseURL, "/")
	if s.Lang == "" {
		s.Lang = "en"
	}
	return s
}

// LoadFolder parses a _meta.yaml file.
func LoadFolder(data []byte, diags *diag.List) Folder {
	var f Folder
	root := decodeStrict(data, 0, &f, diags)
	validateDefaults(root, []string{"defaults"}, f.Defaults, 0, diags)
	if f.Sort != "" && f.Sort != "title" && f.Sort != "date" {
		diags.Errorf(keyLine(root, 0, "sort"), 0, diag.CodeConfigValue, "Valores aceitos: title, date.", "sort desconhecido %q", f.Sort)
	}
	return f
}

// LoadFrontmatter parses a page frontmatter. lineOffset is the number of
// lines before the YAML content in the page file.
func LoadFrontmatter(data []byte, lineOffset int, diags *diag.List) Frontmatter {
	var fm Frontmatter
	root := decodeStrict(data, lineOffset, &fm, diags)
	validateDefaults(root, nil, fm.Defaults, lineOffset, diags)
	if fm.Date != nil {
		if _, err := ParseDate(*fm.Date); err != nil {
			diags.Errorf(keyLine(root, lineOffset, "date"), 0, diag.CodeDate, "Use o formato AAAA-MM-DD, por exemplo date: 2026-03-14.", "data inválida %q", *fm.Date)
		}
	}
	if fm.Title != nil && strings.TrimSpace(*fm.Title) == "" {
		diags.Errorf(keyLine(root, lineOffset, "title"), 0, diag.CodeFrontmatter, "Remova a chave ou preencha o título.", "'title' está vazio")
	}
	return fm
}

// SplitFrontmatter separates a leading '---' YAML block from the body.
// It returns the YAML bytes, the body and the number of lines that precede
// the body (so body line N is file line N+bodyOffset).
func SplitFrontmatter(src []byte, diags *diag.List) (fm []byte, body []byte, yamlOffset, bodyOffset int) {
	first, rest, more := cutLine(src)
	if !more || strings.TrimRight(string(first), " \t\r") != "---" {
		return nil, src, 0, 0
	}
	start := len(src) - len(rest)
	lineNo := 1
	cur := rest
	for {
		line, after, more := cutLine(cur)
		lineNo++
		trimmed := strings.TrimRight(string(line), " \t\r")
		if trimmed == "---" || trimmed == "..." {
			// YAML line 1 is file line 2; body line 1 is file line lineNo+1.
			return src[start : len(src)-len(cur)], after, 1, lineNo
		}
		if !more {
			break
		}
		cur = after
	}
	diags.Errorf(1, 1, diag.CodeFrontmatter, "Feche o frontmatter com uma linha contendo apenas '---'.", "frontmatter aberto com '---' e nunca fechado")
	return nil, src, 0, 0
}

func cutLine(b []byte) (line, rest []byte, more bool) {
	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		return b[:i], b[i+1:], true
	}
	return b, nil, false
}

var yamlErrLine = regexp.MustCompile(`^line (\d+): (.*)$`)
var unknownField = regexp.MustCompile(`^field (\S+) not found in type (\S+)$`)
var anyLine = regexp.MustCompile(`line (\d+)`)

// decodeStrict decodes data into v, rejecting unknown keys, and reports
// errors as diagnostics. It returns the YAML node tree for position lookups.
func decodeStrict(data []byte, lineOffset int, v any, diags *diag.List) *yaml.Node {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		line := 0
		msg := strings.TrimPrefix(err.Error(), "yaml: ")
		if m := anyLine.FindStringSubmatch(msg); m != nil {
			line, _ = strconv.Atoi(m[1])
			line += lineOffset
		}
		diags.Errorf(line, 0, diag.CodeConfigInvalid, "Verifique indentação, dois-pontos e aspas.", "YAML inválido: %s", msg)
		return &root
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	err := dec.Decode(v)
	if err == nil || errors.Is(err, io.EOF) {
		return &root
	}
	var te *yaml.TypeError
	if !errors.As(err, &te) {
		diags.Errorf(0, 0, diag.CodeConfigInvalid, "", "YAML inválido: %v", err)
		return &root
	}
	for _, e := range te.Errors {
		line := 0
		msg := e
		if m := yamlErrLine.FindStringSubmatch(e); m != nil {
			line, _ = strconv.Atoi(m[1])
			line += lineOffset
			msg = m[2]
		}
		if m := unknownField.FindStringSubmatch(msg); m != nil {
			known := knownKeys(v, m[2])
			hint := "Chaves válidas: " + strings.Join(known, ", ") + "."
			if alt, ok := synonyms[m[1]]; ok && contains(known, alt.key) {
				hint = alt.hint + " " + hint
			} else if s := diag.Suggest(m[1], known); s != "" {
				hint = fmt.Sprintf("Você quis dizer '%s'? %s", s, hint)
			}
			diags.Errorf(line, 0, diag.CodeConfigUnknownKey, hint, "chave desconhecida '%s'", m[1])
			continue
		}
		diags.Errorf(line, 0, diag.CodeConfigValue, "", "valor inválido: %s", translateYAML(msg, v))
	}
	return &root
}

// synonyms maps keys other tools use (Docusaurus, Mintlify, Hugo, VitePress)
// to Wavemark's, so the hint names the right key even when it looks nothing
// like the typo.
var synonyms = map[string]struct{ key, hint string }{
	"title":            {"label", "Itens de nav usam 'label' em vez de 'title'."},
	"text":             {"label", "Itens de nav usam 'label' em vez de 'text'."},
	"name":             {"label", "Use 'label'."},
	"link":             {"href", "Use 'href'."},
	"url":              {"href", "Use 'href'."},
	"to":               {"href", "Use 'href'."},
	"siteUrl":          {"baseURL", "Use 'baseURL'."},
	"site_url":         {"baseURL", "Use 'baseURL'."},
	"language":         {"lang", "Use 'lang'."},
	"locale":           {"lang", "Use 'lang'."},
	"sidebar_position": {"order", "Use 'order'."},
	"weight":           {"order", "Use 'order'."},
	"position":         {"order", "Use 'order'."},
	"summary":          {"description", "Use 'description'."},
	"excerpt":          {"description", "Use 'description'."},
	"subtitle":         {"description", "Use 'description'."},
	"authors":          {"author", "Use 'author' com um texto (\"Ana e Rui\")."},
	"published":        {"date", "Use 'date: AAAA-MM-DD'."},
	"pubDate":          {"date", "Use 'date: AAAA-MM-DD'."},
	"draft":            {"hidden", "Não há rascunhos; 'hidden: true' tira a página da navegação."},
	"sortBy":           {"sort", "Use 'sort: date' ou 'sort: title'."},
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

var unmarshalRe = regexp.MustCompile("^cannot unmarshal !!(\\w+) `(.*)` into (\\S+)$")

// translateYAML rewrites yaml.v3 type errors without Go type names.
func translateYAML(msg string, v any) string {
	m := unmarshalRe.FindStringSubmatch(msg)
	if m == nil {
		return msg
	}
	got := map[string]string{"str": "um texto", "int": "um número", "bool": "um booleano", "seq": "uma lista", "map": "um mapa", "float": "um número"}[m[1]]
	if got == "" {
		got = m[1]
	}
	var want string
	switch t := m[3]; {
	case t == "string":
		want = "um texto"
	case t == "int":
		want = "um número inteiro"
	case t == "bool":
		want = "true ou false"
	case t == "[]string":
		want = "uma lista de textos, como [a, b]"
	case strings.HasPrefix(t, "[]"):
		want = "uma lista"
	default:
		if keys := knownKeys(v, t); len(keys) > 0 {
			want = "um mapa com as chaves " + strings.Join(keys, ", ")
		} else {
			want = "outro tipo"
		}
	}
	return fmt.Sprintf("esperava %s, recebeu %s", want, got)
}

// knownKeys lists the YAML keys of the struct type named typeName reachable
// from v (yaml.v3 reports the innermost struct type).
func knownKeys(v any, typeName string) []string {
	var found []string
	seen := map[reflect.Type]bool{}
	var walk func(t reflect.Type)
	walk = func(t reflect.Type) {
		for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct || seen[t] {
			return
		}
		seen[t] = true
		if t.String() == typeName {
			found = yamlKeys(t)
			return
		}
		for i := range t.NumField() {
			walk(t.Field(i).Type)
		}
	}
	walk(reflect.TypeOf(v))
	return found
}

func yamlKeys(t reflect.Type) []string {
	var keys []string
	for i := range t.NumField() {
		f := t.Field(i)
		tag := f.Tag.Get("yaml")
		name, opts, _ := strings.Cut(tag, ",")
		if opts == "inline" {
			keys = append(keys, yamlKeys(f.Type)...)
			continue
		}
		if name != "" && name != "-" {
			keys = append(keys, name)
		}
	}
	return keys
}

// Fields documents the keys of a config struct for `wavemark spec`.
type Field struct {
	Key, Type, Doc string
}

// Describe lists the documented keys of one of the config types.
func Describe(v any) []Field {
	var out []Field
	var walk func(t reflect.Type)
	walk = func(t reflect.Type) {
		for i := range t.NumField() {
			f := t.Field(i)
			name, opts, _ := strings.Cut(f.Tag.Get("yaml"), ",")
			if opts == "inline" {
				walk(f.Type)
				continue
			}
			out = append(out, Field{Key: name, Type: typeName(f.Type), Doc: f.Tag.Get("doc")})
		}
	}
	walk(reflect.TypeOf(v))
	return out
}

func typeName(t reflect.Type) string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return "texto"
	case reflect.Bool:
		return "booleano"
	case reflect.Int:
		return "inteiro"
	case reflect.Slice:
		if t.Elem().Kind() == reflect.String {
			return "lista de textos"
		}
		return "lista"
	case reflect.Struct:
		return "mapa"
	}
	return t.String()
}

func validateDefaults(root *yaml.Node, path []string, d Defaults, lineOffset int, diags *diag.List) {
	if d.Layout == nil {
		return
	}
	for _, l := range Layouts {
		if *d.Layout == l {
			return
		}
	}
	line := keyLine(root, lineOffset, append(path, "layout")...)
	hint := "Valores aceitos: " + strings.Join(Layouts, ", ") + "."
	if s := diag.Suggest(*d.Layout, Layouts); s != "" {
		hint = fmt.Sprintf("Você quis dizer '%s'? %s", s, hint)
	}
	diags.Errorf(line, 0, diag.CodeConfigValue, hint, "layout desconhecido '%s'", *d.Layout)
}

// keyLine finds the line of a key path in a YAML document (0 if absent).
// Path elements are map keys or decimal sequence indexes.
func keyLine(root *yaml.Node, lineOffset int, path ...string) int {
	n := root
	if n.Kind == yaml.DocumentNode && len(n.Content) > 0 {
		n = n.Content[0]
	}
	line := 0
	for _, p := range path {
		switch n.Kind {
		case yaml.MappingNode:
			var next *yaml.Node
			for i := 0; i+1 < len(n.Content); i += 2 {
				if n.Content[i].Value == p {
					line = n.Content[i].Line
					next = n.Content[i+1]
					break
				}
			}
			if next == nil {
				return lineOr(line, lineOffset)
			}
			n = next
		case yaml.SequenceNode:
			idx, err := strconv.Atoi(p)
			if err != nil || idx >= len(n.Content) {
				return lineOr(line, lineOffset)
			}
			n = n.Content[idx]
			line = n.Line
		default:
			return lineOr(line, lineOffset)
		}
	}
	return lineOr(line, lineOffset)
}

func lineOr(line, offset int) int {
	if line == 0 {
		return 0
	}
	return line + offset
}
