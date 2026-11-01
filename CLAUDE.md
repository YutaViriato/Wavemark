# Wavemark

Wavemark é um runtime que transforma uma pasta de arquivos Markdown em um site. Você aponta o binário para uma pasta e o servidor renderiza as páginas; editar um `.md` muda o site sem etapa de build.

Os autores do conteúdo são principalmente agentes de IA. Isso orienta quase todas as decisões abaixo: gramática estrita, validação com schema, erros acionáveis e uma especificação legível por máquina.

## Decisões de produto

| Tema | Decisão |
|---|---|
| Modelo de execução | Runtime de verdade: servidor HTTP lê a pasta, mantém um snapshot em memória e reconstrói quando arquivos mudam (fsnotify). Sem build estático por enquanto. |
| Público | Documentação e notas longas, com componentes de site (cards, colunas, hero, botões) para páginas de entrada. |
| Sintaxe de componentes | Tags estilo JSX, declarativas, **sem executar JS**: `<Card title="X" href="/y">`. Nome com inicial maiúscula = componente. Valores de atributo: `"string"`, `'string'`, `{literal JSON}` ou atributo booleano sem valor. Expressões JS são erro. |
| Regras da gramática | Tags de componente ficam em linhas próprias (abertura, fechamento, autofechada). Exceção: forma de linha única `<Step title="x">texto</Step>`. A tag de abertura pode ocupar várias linhas. O conteúdo entre tags é Markdown e pode ser indentado (o componente remove a indentação). |
| HTML cru | Proibido. Vira erro no `check` e não é renderizado. Tudo passa por componentes para que o agente não quebre a identidade visual. |
| Blocos de código indentados | Desligados. Só blocos cercados (```` ``` ````). Isso permite indentar conteúdo dentro de componentes sem gerar código por acidente. |
| Configuração | Cascata em três níveis: `wavemark.yaml` (raiz, chave `defaults`) → `_meta.yaml` por pasta (chave `defaults`) → frontmatter da página. Mapas mesclam em profundidade, escalares sobrescrevem, listas substituem. Chaves desconhecidas são erro (pega `tittle:`). Chaves de site (`nav`, `title` do site, `logo`) só existem em `wavemark.yaml`. |
| Título da página | `title` do frontmatter ou, na falta dele, o primeiro `# H1` (que sai do corpo e vai para o layout). H1 no corpo junto com `title` no frontmatter é erro. Mais de um H1 é erro. |
| Rotas | Derivadas das pastas: `guia/instalacao.md` → `/guia/instalacao`; `index.md` ou `README.md` → rota da pasta. A fonte de cada página também é servida em `<rota>.md`, e o site expõe `/llms.txt`. |
| Links internos | Links para `.md` relativos são reescritos para a rota e validados no índice global, incluindo âncoras `#id`. Link quebrado é erro. |
| Página inválida | `wavemark serve` só sobe se o `check` passar. Com o servidor rodando, uma página que fica inválida continua servindo a última versão válida, e o erro vai para o log. `wavemark dev` mostra o erro sobreposto na página e recarrega via SSE. |
| Layouts | `docs` (padrão: sidebar, conteúdo, sumário à direita, fundo neutro `#fcfcfb` como platform.claude.com/docs) e `landing` (hero, títulos grandes, fundo marfim `#faf9f5` como claude.com). |
| Componentes v1 | `Callout` (+ atalhos `Note`, `Tip`, `Info`, `Warning`, `Danger`), `Card`, `Columns`, `Tabs`/`Tab`, `CodeGroup`, `Steps`/`Step`, `Accordion`, `Button`, `Hero`, `Figure`. Tabelas GFM. Badge e componentes inline, ícones, Mermaid, matemática e busca ficam para depois. |
| JS no cliente | Vanilla, mínimo, embutido no binário: abas, copiar código, troca de tema, live reload no `dev`. O site continua legível sem JS. |
| Distribuição | Fora do escopo por enquanto. |

## Identidade visual

A referência são os sites atuais (medidos em novembro de 2026), não as diretrizes públicas antigas (Poppins/Lora). Valores medidos e metodologia ficam em `design/reference.md`.

- As fontes Anthropic Sans/Serif/Mono são proprietárias. **Nunca embutir nem fazer hotlink delas.** Usamos alternativas livres (OFL) escolhidas por comparação visual: Hanken Grotesk (sans), Newsreader (serif) e JetBrains Mono. A config permite apontar arquivos de fonte próprios.
- Não distribuir o logo da Claude nem o wordmark da Anthropic. O logo vem da config do site.
- Todo valor visual vem de tokens CSS (`--wm-*`). Componentes não usam cores literais.
- Verificação visual: `tools/visual/` (Playwright) captura as páginas do Wavemark e as de referência no mesmo viewport e compara estilos computados. Comparação por pixel não serve porque as fontes diferem.

## Para agentes que escrevem conteúdo

- `wavemark spec` imprime a referência completa de componentes e frontmatter, gerada dos mesmos schemas que validam. Leia antes de escrever.
- `wavemark check [dir] --json` devolve diagnósticos com arquivo, linha, coluna, código e sugestão. Rode até ficar limpo.

## Arquitetura (Go)

```
cmd/wavemark/        CLI: serve, dev, check, spec
internal/diag/       diagnósticos (arquivo, linha, coluna, código, mensagem, dica)
internal/config/     wavemark.yaml, _meta.yaml, frontmatter, cascata e validação de chaves
internal/markdown/   goldmark + parser de componentes + renderers (código, títulos, links)
internal/components/ schemas e renderização de cada componente
internal/site/       varredura da pasta, rotas, navegação, índice global, validação de links
internal/server/     HTTP, snapshot, watcher, last-known-good, live reload
internal/theme/      templates, CSS, JS e fontes embutidos (embed.FS)
site/                site de exemplo: a documentação do próprio Wavemark, escrita em Wavemark
design/              referência visual medida
tools/visual/        scripts Playwright de comparação visual
```

Dependências: `github.com/yuin/goldmark`, `github.com/alecthomas/chroma/v2` (realce no servidor, por classes CSS), `github.com/fsnotify/fsnotify`, `gopkg.in/yaml.v3`. Evitar dependências novas sem motivo.

## Comandos

```bash
go build -o bin/wavemark ./cmd/wavemark
./bin/wavemark dev site          # servidor de desenvolvimento em :4321
./bin/wavemark check site --json # validação
./bin/wavemark spec              # referência para agentes
go vet ./...

# Verificação visual (com o dev rodando em :4321)
cd tools/visual && npm install && npx playwright install chromium
node compare.mjs --scheme light   # estilos computados vs. platform.claude.com e claude.com
node compare.mjs --scheme dark
node shots.mjs --pages /,/guia/instalacao --scheme light,dark --scroll 0,900   # capturas em tools/visual/shots/
```

## Estado atual (v1)

- Implementado: parser de componentes, validação com diagnósticos, cascata de config, rotas, navegação, links e âncoras validados, tema `docs` e `landing` (claro e escuro), runtime com snapshot atômico, watcher, SSE, última versão válida, `llms.txt`, fonte `.md` de cada página, CSP.
- `compare.mjs`: 53/53 propriedades iguais à referência em claro e escuro (fontes excluídas por serem proprietárias).
- Desempenho: 601 páginas pesadas reconstroem em ~0,5 s em 2 núcleos. O realce de código é memoizado por hash e a sidebar é renderizada uma vez por snapshot.
- Ainda não existe: testes automatizados, busca, ícones, componentes inline (Badge), Mermaid, matemática, build estático, i18n além de en/pt.

## Armadilhas conhecidas

- O goldmark chama `Continue` de fora para dentro em cada linha. Por isso a tag de fechamento é decidida em `componentParser.owns`, que olha `pc.OpenedBlocks()`. Mexa ali com cuidado e teste aninhamento, cercas de código com `</Tag>` dentro e componentes autofechados.
- O renderer de tabela do goldmark usa `style=""` por padrão. Mantenha `TableCellAlignAttribute`, porque a CSP bloqueia estilos inline.
- Não use `style=` em nenhum HTML gerado, pelo mesmo motivo; use classes (`wm-columns--3`).
- Ao matar servidores em scripts, use `pgrep -f "^./bin/wavemark"`. Com `pkill -f "wavemark dev"`, o padrão casa com o próprio shell.

## Convenções

- Go idiomático, `gofmt`, erros com contexto (`fmt.Errorf("...: %w", err)`).
- Saída HTML sempre escapada; atributos de URL validados (só http(s), mailto, relativo, `/` e `#`).
- Arquivos servidos só a partir do índice da varredura, nunca por junção de caminho da requisição. Symlinks não são seguidos.
- Documentação e mensagens para o usuário em português; identificadores e código em inglês.
