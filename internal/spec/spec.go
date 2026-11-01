// Package spec generates the authoring reference printed by `wavemark spec`.
// It is built from the same schemas that validate content, so it cannot
// drift from the implementation.
package spec

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/components"
	"github.com/YutaViriato/Wavemark/internal/config"
)

// Markdown returns the full reference as Markdown.
func Markdown() string {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }

	w("# Referência do Wavemark para autores\n\n")
	w("Wavemark transforma uma pasta de Markdown em site. Este documento é a referência completa da sintaxe aceita. ")
	w("Valide sempre com `wavemark check <pasta> --json` e corrija até não haver erros.\n\n")

	w("## Estrutura da pasta\n\n")
	w("- Cada arquivo `.md` é uma página. A rota vem do caminho: `guia/instalacao.md` → `/guia/instalacao`.\n")
	w("- `index.md` (ou `README.md`) é a página da própria pasta: `guia/index.md` → `/guia`. A raiz `index.md` é a home.\n")
	w("- Cada pasta é uma seção da navegação lateral. `_meta.yaml` na pasta define título, ordem e configurações herdadas.\n")
	w("- `wavemark.yaml` na raiz configura o site.\n")
	w("- Outros arquivos (imagens, PDFs) são servidos no mesmo caminho: `img/a.png` → `/img/a.png`.\n")
	w("- Nomes que começam com `.` e `node_modules` são ignorados.\n\n")

	w("## Markdown aceito\n\n")
	w("CommonMark com extensões GFM: tabelas, ~~riscado~~, listas de tarefas, autolinks e notas de rodapé (`[^1]`).\n\n")
	w("Regras do Wavemark:\n\n")
	w("- **HTML cru é proibido** (inclusive `<br>`, `<div>`, `<img>`). Use Markdown ou componentes. Comentários `<!-- -->` são permitidos e não aparecem.\n")
	w("- **Só blocos de código cercados** (```` ``` ````). Blocos indentados com 4 espaços não existem; indentação vira texto normal.\n")
	w("- A cerca aceita linguagem e título: ```` ```python title=\"main.py\" ````.\n")
	w("- **Um título por página**: `title` no frontmatter ou o primeiro `# H1` do corpo, nunca os dois. Use `##` e `###` para seções; eles formam o sumário.\n")
	w("- Âncoras são geradas dos títulos sem acentos: `## Configuração avançada` → `#configuracao-avancada`. Para fixar uma âncora: `## Título {#minha-ancora}`.\n")
	w("- **Links internos apontam para o arquivo**: `[Instalação](./instalacao.md)`, `[Seção](../guia/x.md#secao)`. Rotas absolutas também valem: `/guia/x`. Links quebrados e âncoras inexistentes são erro.\n")
	w("- Imagens: `![descrição](./img/a.png)`. O arquivo precisa existir.\n\n")

	w("## Componentes\n\n")
	w("Componentes são tags com nome em PascalCase, escritas como JSX declarativo. Não há JavaScript.\n\n")
	w("Regras de sintaxe:\n\n")
	w("1. Cada tag fica **em uma linha própria**: abertura, fechamento e tags autofechadas (`<Button ... />`).\n")
	w("2. A tag de abertura pode quebrar linhas entre atributos, mas não pode conter linha em branco.\n")
	w("3. O conteúdo entre as tags é Markdown e pode ser indentado; a indentação da primeira linha de conteúdo é removida.\n")
	w("4. Forma curta em uma linha: `<Step title=\"Instale\">Rode o instalador.</Step>`.\n")
	w("5. Valores: texto entre aspas `title=\"...\"` (escape aspas com `\\\"`), literais JSON entre chaves `cols={3}` `open={false}`, ou o nome sozinho para verdadeiro (`open`).\n")
	w("6. Atributos desconhecidos, obrigatórios ausentes e valores fora do tipo são erro.\n\n")

	for _, s := range components.All() {
		w("### <%s>\n\n%s\n\n", s.Name, s.Summary)
		if len(s.Layouts) > 0 {
			w("Layouts: %s.\n\n", strings.Join(s.Layouts, ", "))
		}
		if s.Parent != "" {
			w("Precisa estar diretamente dentro de `<%s>`.\n\n", s.Parent)
		}
		switch s.Children {
		case components.ChildNone:
			w("Sem conteúdo: use a forma autofechada `<%s ... />`.\n\n", s.Name)
		case components.ChildOnly:
			var allowed []string
			for _, a := range s.Allowed {
				if a == components.ChildCode {
					allowed = append(allowed, "blocos de código cercados")
				} else {
					allowed = append(allowed, "`<"+a+">`")
				}
			}
			w("Conteúdo: apenas %s.\n\n", strings.Join(allowed, " ou "))
		default:
			w("Conteúdo: Markdown e outros componentes.\n\n")
		}
		if len(s.Attrs) > 0 {
			w("| Atributo | Tipo | Obrigatório | Padrão | Descrição |\n|---|---|---|---|---|\n")
			for _, a := range s.Attrs {
				typ := a.Type.String()
				if a.Type == components.TEnum {
					typ = strings.Join(quoteAll(a.Enum), " \\| ")
				}
				req := "não"
				if a.Required {
					req = "sim"
				}
				def := ""
				if a.Default != nil {
					def = "`" + jsonish(a.Default) + "`"
				}
				w("| `%s` | %s | %s | %s | %s |\n", a.Name, typ, req, def, a.Doc)
			}
			w("\n")
		}
		w("```mdx\n%s\n```\n\n", s.Example)
	}

	w("## Frontmatter da página\n\n")
	w("Bloco YAML no topo do arquivo, entre linhas `---`. Todas as chaves são opcionais.\n\n")
	fieldTable(&b, config.Describe(config.Frontmatter{}))
	w("```yaml\n---\ntitle: Instalação\ndescription: Como instalar o Wavemark em um minuto.\norder: 1\n---\n```\n\n")

	w("## _meta.yaml (por pasta)\n\n")
	fieldTable(&b, config.Describe(config.Folder{}))
	w("Chaves de `defaults`:\n\n")
	fieldTable(&b, config.Describe(config.Defaults{}))
	w("```yaml\ntitle: Guia\norder: [instalacao, configuracao, deploy]\ndefaults:\n  toc: true\n```\n\n")

	w("## wavemark.yaml (raiz)\n\n")
	fieldTable(&b, config.Describe(config.Site{}))
	w("```yaml\ntitle: Acme Docs\ndescription: Documentação da Acme.\nlang: pt-BR\nnav:\n  - label: Guia\n    href: /guia\n  - label: API\n    href: ./api/index.md\ndefaults:\n  layout: docs\n```\n\n")

	w("## Cascata de configuração\n\n")
	w("`wavemark.yaml › defaults` → `_meta.yaml › defaults` de cada pasta, da mais externa para a mais interna → frontmatter. ")
	w("O nível mais específico vence. Listas substituem, não somam.\n\n")

	w("## Layouts\n\n")
	w("- `docs` (padrão): navegação lateral, título serifado, sumário à direita, páginas anterior/próxima.\n")
	w("- `landing`: página de entrada sem navegação lateral. Comece com `<Hero>`; use `##` para seções e `<Columns>` com `<Card>` para grades.\n")
	return b.String()
}

func fieldTable(b *strings.Builder, fields []config.Field) {
	b.WriteString("| Chave | Tipo | Descrição |\n|---|---|---|\n")
	for _, f := range fields {
		fmt.Fprintf(b, "| `%s` | %s | %s |\n", f.Key, f.Type, f.Doc)
	}
	b.WriteString("\n")
}

func quoteAll(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = `"` + s + `"`
	}
	return out
}

func jsonish(v any) string {
	if s, ok := v.(string); ok {
		return `"` + s + `"`
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// JSON returns the component schemas as JSON, for tools.
func JSON() ([]byte, error) {
	type attr struct {
		Name     string   `json:"name"`
		Type     string   `json:"type"`
		Required bool     `json:"required"`
		Enum     []string `json:"enum,omitempty"`
		Default  any      `json:"default,omitempty"`
		Doc      string   `json:"doc"`
	}
	type comp struct {
		Name     string   `json:"name"`
		Summary  string   `json:"summary"`
		AliasOf  string   `json:"aliasOf,omitempty"`
		Parent   string   `json:"parent,omitempty"`
		Layouts  []string `json:"layouts,omitempty"`
		Children string   `json:"children"`
		Allowed  []string `json:"allowed,omitempty"`
		Attrs    []attr   `json:"attributes"`
		Example  string   `json:"example"`
	}
	var out []comp
	for _, s := range components.All() {
		c := comp{Name: s.Name, Summary: s.Summary, AliasOf: s.AliasOf, Parent: s.Parent, Layouts: s.Layouts, Example: s.Example, Attrs: []attr{}}
		switch s.Children {
		case components.ChildNone:
			c.Children = "none"
		case components.ChildOnly:
			c.Children = "only"
			c.Allowed = s.Allowed
		default:
			c.Children = "markdown"
		}
		for _, a := range s.Attrs {
			c.Attrs = append(c.Attrs, attr{Name: a.Name, Type: a.Type.String(), Required: a.Required, Enum: a.Enum, Default: a.Default, Doc: a.Doc})
		}
		out = append(out, c)
	}
	return json.MarshalIndent(map[string]any{"components": out}, "", "  ")
}
