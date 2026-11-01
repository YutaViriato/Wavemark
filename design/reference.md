# Referência visual

Valores medidos em 1º de novembro de 2026 com Chromium headless (Playwright), viewport 1440×900, a partir dos estilos computados de:

- `platform.claude.com/docs` (layout `docs`)
- `www.claude.com` e `www.anthropic.com` (layout `landing`)

Os scripts de medição estão em `tools/visual/`. Repita a medição antes de mudar tokens: os sites mudam.

## Tipografia

Os três sites usam fontes próprias (Anthropic Sans, Anthropic Serif, Anthropic Mono). Elas são proprietárias e não podem ser embutidas. Comparamos candidatas livres lado a lado com as capturas:

| Papel | Escolha (OFL) | Por quê | Descartadas |
|---|---|---|---|
| Sans | Hanken Grotesk | Grotesca com métrica e ritmo mais próximos; no parágrafo de corpo a 832px quebra as linhas nos mesmos pontos que a referência | Inter, Geist, Onest e Public Sans (largas demais), Manrope (geométrica demais). Instrument Sans e Schibsted Grotesk não carregaram no teste e ficaram sem avaliação |
| Serif | Newsreader | Serif de texto contemporânea com eixo `opsz`; contraste e terminais próximos em 26px e em 56px | Source Serif 4 (pesada no display), Literata e Noto Serif (largas), Gelasio/Spectral/Caslon (clássicas demais) |
| Mono | JetBrains Mono | O próprio anthropic.com carrega JetBrains Mono como fonte de código | — |

Licenças em `internal/theme/assets/fonts/OFL-*.txt`.

## Layout `docs` (platform.claude.com/docs)

| Elemento | Claro | Escuro | Métrica |
|---|---|---|---|
| Fundo da página | `#fcfcfb` | `#151515` | |
| Texto forte (títulos, nav ativa) | `#0b0b0b` | `#f0efec` | |
| Texto de corpo | `#52514e` | `#c3c2b7` | sans 15px / 22px |
| Texto apagado (breadcrumb, TOC, nav inativa) | `#898781` | `#898781` | 14px / 20px |
| Bordas | `rgba(11,11,11,.1)` | `rgba(255,255,255,.1)` | 1px |
| Superfície realçada (ativo, th) | `rgba(11,11,11,.05–.1)` | `rgba(255,255,255,.05–.15)` | |
| Link | `#184f95`, sublinhado `rgba(24,79,149,.4)`, offset 3px | `#6da7ec` | |
| Código inline | `#8e2626` sobre `rgba(11,11,11,.05)`, borda `.1` | `#ec7e7e` sobre `rgba(255,255,255,.05)` | mono 13.5px, raio 5.4px, padding 0.84px 3.4px |
| Bloco de código | fundo `#fff`, borda `.1`, raio 11px | fundo `#1a1a19` | mono 13px / 19px, cabeçalho com abas 14px peso 500 |
| Realce de sintaxe | paleta GitHub Primer: texto `#1f2328`, keyword `#cf222e`, string `#0a3069`, número/builtin `#0550ae`, parâmetro `#953800` | paleta estilo VS Code Dark+: `#9cdcfe`, `#ce9178`, `#b5cea8`, `#dcdcaa` | |

Estrutura:

- Header: 56px de altura, borda inferior, padding horizontal 16px. Logo em serif ~20px. Nav superior com itens em pílula de 32px de altura; item ativo com fundo `rgba(11,11,11,.1)`.
- Sidebar: 264px, borda direita. Itens de 32px de altura, raio 8px, padding horizontal 8px, 14px. Item ativo com fundo `rgba(11,11,11,.1)` e texto forte. Grupo: 14px peso 500, texto forte.
- Conteúdo: começa 45px depois da sidebar, largura 832px.
  - Breadcrumb: 14px apagado.
  - H1: serif 26.25px peso 520, line-height 1.25.
  - Lead: sans 16.875px / 1.5, texto de corpo.
  - H2: serif 20.625px peso 520. H3: serif 16.875px peso 520.
  - Strong: peso 600, texto forte.
- TOC: 224px à direita, padding 32px 0. Itens de 28px com padding-left 12px. Ativo com peso 500, texto forte e borda esquerda de 2px. Nível 3 recuado 12px.
- Tabela: borda 1px, raio 7px. th com fundo `.05`, peso 500, padding 6.5px 11.25px.
- Callout: raio 11px, padding 10px 14px, 14px / 20px, ícone à esquerda.
- Cards da home: borda 1px, raio ~12px, fundo `#fff`.

## Layout `landing` (claude.com / anthropic.com)

| Elemento | Valor |
|---|---|
| Fundo | `#faf9f5` (marfim). Escuro no claude.com: `#141413`, texto `#faf9f5` |
| Texto | `#141413`; secundário `#5e5d59` |
| H1 do hero | serif 56px peso 400, line-height 1.15 |
| Subtítulo do hero | serif 18px |
| H2 de seção | serif 52px peso 500 |
| Botão primário | fundo `#141413`, texto `#faf9f5`, raio 8px, 15px peso 500 |
| Botão secundário | borda 1px `#d1cfc5`, fundo transparente, raio 8px |
| Card | fundo `#fff` (escuro: `#1f1e1d`), borda `#e8e6dc`, raio 16–32px |
| Acento | clay `#d97757`, interativo `#c6613f` |

## Escala de cinzas quentes (anthropic.com e claude.com)

`000 #fff`, `050 #faf9f5`, `100 #f5f4ed`, `150 #f0eee6`, `200 #e8e6dc`, `250 #dedcd1`, `300 #d1cfc5`, `350 #c2c0b6`, `400 #b0aea5`, `450 #9c9a92`, `500 #87867f`, `550 #73726c`, `600 #5e5d59`, `650 #4d4c48`, `700 #3d3d3a`, `750 #30302e`, `800 #262624`, `850 #1f1e1d`, `900 #1a1918`, `950 #141413`, `1000 #000`.

Acentos: clay `#d97757` (interativo `#c6613f`), olive `#788c5d`, sky `#6a9bcc`, fig `#c46686`, cactus `#bcd1ca`, heather `#cbcadb`, oat `#e3dacc`, erro `#bf4d43`.

Espaçamento: 4, 8, 12, 16, 24, 32, 40, 48, 64, 80, 96, 128. Raios: 4, 8, 12, 16. Easing: `cubic-bezier(.165,.84,.44,1)` (out-quart).

## Layout `article` (anthropic.com/research e /news)

Medido em `anthropic.com/research/claude-think-tool` e `anthropic.com/news/2026-usage-policy-update`:

- Título centralizado: sans peso 700, 32.5px; subtítulo serif 32px logo abaixo.
- Três colunas a 1440px: capítulos numerados (`01`, `02`… em mono 12px) em x=80 com 198px de largura e bordas entre itens; corpo de 625px centralizado (x=400); fatos (Publicado, Equipe, Tópicos) em x=1148 com 197px.
- Corpo: serif 16.7px / 25.5px (notícias) ou sans (pesquisa), cor `#141413`. H2 sans 700 26px, margem 48px acima e 24px abaixo. Parágrafos separados por 24px. Links sublinhados na cor do texto.

## Movimento (anthropic.com, claude.com, platform.claude.com)

Extraído dos `@keyframes` e `transition` do CSS dos três sites:

- Curvas: `cubic-bezier(.165,.84,.44,1)` (out-quart) e `cubic-bezier(.19,1,.22,1)` (out-expo).
- Durações: 0.15 a 0.3s em hover e cor; 0.5 a 0.8s em entradas.
- Entradas: `word-rise` (translateY .42em), `unblur` (blur 4px → 0) e `fade-in` no anthropic.com; `raise` (translateY 16px) e `fadeUp` no claude.com; `reveal-in` (translateY 4px) na documentação.
- A documentação usa view transitions e `animation-timeline`; o anthropic.com anima traços com `stroke-dash`.
- Nada pisca nem fica em loop no conteúdo de leitura.

## Ilustrações

Os ícones de platform.claude.com/docs e claude.com são traço preto grosso (~2px a 48px), cantos levemente irregulares e preenchimento chapado quente (`#e3dacc` oat, `#dedcd1`). As peças grandes do anthropic.com são gravuras e colagens com textura. O Wavemark gera arte própria com o mesmo vocabulário (traço de caneta e preenchimento oat fora de registro), sem copiar nenhuma peça.
