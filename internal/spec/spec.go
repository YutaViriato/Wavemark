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
	"github.com/YutaViriato/Wavemark/internal/sketch"
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
	w("- A cerca aceita linguagem e título: ```` ```python title=\"main.py\" ````. Há realce para mais de 200 linguagens (bash, python, typescript, go, rust, json, yaml, sql, html, css, diff…); `text` desliga o realce.\n")
	w("- Blocos `mermaid` e de matemática ainda não são renderizados: aparecem como código. Para diagramas, use uma imagem SVG com `<Figure>`.\n")
	w("- **Um título por página**: `title` no frontmatter ou o primeiro `# H1` do corpo, nunca os dois. Use `##` e `###` para seções; eles formam o sumário.\n")
	w("- Âncoras são geradas dos títulos sem acentos: `## Configuração avançada` → `#configuracao-avancada`. Para fixar uma âncora: `## Título {#minha-ancora}`.\n")
	w("- Destinos aceitos em links, `href` e `nav`: arquivo `.md` relativo, rota absoluta (`/guia`), âncora (`#secao`), `https://`, `mailto:` e `tel:`.\n")
	w("- **Links internos apontam para o arquivo**: `[Instalação](./instalacao.md)`, `[Seção](../guia/x.md#secao)`. Rotas absolutas também valem: `/guia/x`. Links quebrados e âncoras inexistentes são erro.\n")
	w("- Imagens: `![descrição](./img/a.png)`. O arquivo precisa existir e a descrição não pode ficar vazia.\n")
	w("- Notas de rodapé: `texto[^1]` e, em uma linha própria, `[^1]: a nota`.\n\n")

	w("## Componentes\n\n")
	w("Componentes são tags com nome em PascalCase, escritas como JSX declarativo. Não há JavaScript.\n\n")
	w("Regras de sintaxe:\n\n")
	w("1. Cada tag fica **em uma linha própria**: abertura, fechamento e tags autofechadas (`<Button ... />`).\n")
	w("2. A tag de abertura pode quebrar linhas entre atributos, mas não pode conter linha em branco.\n")
	w("3. O conteúdo entre as tags é Markdown e pode ser indentado; a indentação da primeira linha de conteúdo é removida.\n")
	w("4. Forma curta em uma linha: `<Step title=\"Instale\">Rode o instalador.</Step>`.\n")
	w("5. Valores: texto entre aspas `title=\"...\"` (escape aspas com `\\\"`), literais JSON entre chaves `cols={3}` `open={false}`, ou o nome sozinho para verdadeiro (`open`).\n")
	w("6. Atributos desconhecidos, obrigatórios ausentes e valores fora do tipo são erro.\n")
	w("7. `<Badge>` e `<Icon />` dentro de um título não entram na âncora: `## Exportar <Badge>Novo</Badge>` gera `#exportar`.\n")
	w("8. **Componentes inline** (`<Badge>`, `<Icon />`) são a exceção: ficam no meio do texto, de títulos ou de células de tabela, e abrem e fecham na mesma linha: `## Exportação <Badge color=\"clay\">Novo</Badge>`.\n\n")

	for _, s := range components.All() {
		w("### <%s>\n\n%s\n\n", s.Name, s.Summary)
		if len(s.Layouts) > 0 {
			w("Layouts: %s.\n\n", strings.Join(s.Layouts, ", "))
		}
		if s.Inline {
			w("Inline: use dentro de um parágrafo, título ou célula de tabela.\n\n")
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
		case components.ChildText:
			w("Conteúdo: texto curto entre as tags.\n\n")
		default:
			if !s.SkipChildren {
				w("Conteúdo: Markdown e outros componentes.\n\n")
			}
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

	w("## Ícones\n\n")
	w("Ícones desenhados à mão, usados em `<Card icon>`, `<Icon name>` e no `icon` do frontmatter. Nomes do Lucide, Font Awesome e Mintlify comuns também funcionam (`rocket-launch`, `settings`, `magnifying-glass`, `envelope`…).\n\n")
	w("%s\n\n", "`"+strings.Join(sketch.Names(), "`, `")+"`")

	w("## Frontmatter da página\n\n")
	w("Bloco YAML no topo do arquivo, entre linhas `---`. Todas as chaves são opcionais.\n\n")
	fieldTable(&b, config.Describe(config.Frontmatter{}))
	w("```yaml\n---\ntitle: Instalação\ndescription: Como instalar o Wavemark em um minuto.\norder: 1\n---\n```\n\n")
	w("Post de blog ou artigo:\n\n```yaml\n---\ntitle: Por que escrevemos devagar\ndescription: Notas sobre o nosso processo de pesquisa.\nlayout: article\ndate: 2026-03-14\nauthor: Ana Lima\n---\n```\n\n")

	w("## _meta.yaml (por pasta)\n\n")
	fieldTable(&b, config.Describe(config.Folder{}))
	w("Chaves de `defaults`:\n\n")
	fieldTable(&b, config.Describe(config.Defaults{}))
	w("```yaml\ntitle: Guia\norder: [instalacao, configuracao, deploy]\ndefaults:\n  toc: true\n```\n\n")
	w("Os `defaults` valem também para o `index.md` da própria pasta. No exemplo abaixo, a página de listagem do blog volta a `layout: docs` no próprio frontmatter.\n\n")
	w("Pasta de blog (posts mais recentes primeiro, todos como artigo):\n\n```yaml\ntitle: Blog\nsort: date\ndefaults:\n  layout: article\n```\n\n")

	w("## wavemark.yaml (raiz)\n\n")
	fieldTable(&b, config.Describe(config.Site{}))
	w("```yaml\ntitle: Acme Docs\ndescription: Documentação da Acme.\nlang: pt-BR\nlogo:\n  text: Acme\n  src: ./img/logo.svg\nnav:\n  - label: Guia\n    href: /guia\n  - label: API\n    href: ./api/index.md\nfooter:\n  text: © 2026 Acme\n  links:\n    - label: GitHub\n      href: https://github.com/acme\ndefaults:\n  layout: docs\n```\n\n")
	w("Não há chaves de tema, cor, CSS ou scripts: a aparência é a do tema.\n\n")

	w("## Cascata de configuração\n\n")
	w("`wavemark.yaml › defaults` → `_meta.yaml › defaults` de cada pasta, da mais externa para a mais interna → frontmatter. ")
	w("O nível mais específico vence. Listas substituem, não somam.\n\n")

	w("## Layouts\n\n")
	w("- `docs` (padrão): navegação lateral, título serifado, sumário à direita, páginas anterior/próxima. Para documentação.\n")
	w("- `article`: leitura longa sem navegação lateral, com capítulos (`##`) à esquerda e data, autor e seção à direita. Para ensaios, posts e notícias.\n")
	w("- `landing`: página de entrada. Comece com `<Hero>`; use `##` para seções, `<Columns>` com `<Card>` ou `<Stat>` para grades e `<Section>` para faixas com fundo. O `title` do frontmatter vira só o título da aba do navegador; o título visível é o do `<Hero>`.\n\n")
	w("## Movimento\n\n")
	w("O tema anima sozinho: o título do `<Hero>` sobe palavra por palavra, ícones e marcadores de `<Steps>` se desenham, blocos aparecem ao rolar e a navegação entre páginas tem transição. Não há nada a configurar. Quem pede menos movimento no sistema vê tudo estático.\n")
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
		Inline   bool     `json:"inline,omitempty"`
		Allowed  []string `json:"allowed,omitempty"`
		Attrs    []attr   `json:"attributes"`
		Example  string   `json:"example"`
	}
	var out []comp
	for _, s := range components.All() {
		c := comp{Name: s.Name, Summary: s.Summary, AliasOf: s.AliasOf, Parent: s.Parent, Layouts: s.Layouts, Inline: s.Inline, Example: s.Example, Attrs: []attr{}}
		switch s.Children {
		case components.ChildNone:
			c.Children = "none"
		case components.ChildOnly:
			c.Children = "only"
			c.Allowed = s.Allowed
		case components.ChildText:
			c.Children = "text"
		default:
			c.Children = "markdown"
		}
		for _, a := range s.Attrs {
			c.Attrs = append(c.Attrs, attr{Name: a.Name, Type: a.Type.String(), Required: a.Required, Enum: a.Enum, Default: a.Default, Doc: a.Doc})
		}
		out = append(out, c)
	}
	return json.MarshalIndent(map[string]any{"components": out, "icons": sketch.Names(), "layouts": config.Layouts}, "", "  ")
}
