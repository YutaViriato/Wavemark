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
| Layouts | `docs` (padrão: sidebar, conteúdo, sumário à direita, fundo neutro `#fcfcfb` como platform.claude.com/docs), `article` (leitura longa como os artigos de anthropic.com/research: capítulos numerados à esquerda, fatos à direita, corpo serifado de 625px) e `landing` (hero, títulos grandes, fundo marfim `#faf9f5` como claude.com). |
| Componentes | Bloco: `Callout` (+ `Note`, `Tip`, `Info`, `Success`/`Check`, `Warning`, `Danger`), `Card` (com `icon`), `Columns` (+ `CardGroup`), `Tabs`/`Tab`, `CodeGroup`, `Steps`/`Step`, `Accordion`, `AccordionGroup`, `Button`, `Hero` (com `highlight` e `image`), `Section`, `Quote`, `Stat`, `PageList`, `Figure`. Inline: `Badge`, `Icon`. Mermaid, matemática e busca ficam para depois. |
| Compatibilidade | Nomes que agentes trazem de Mintlify/Docusaurus viram aliases quando o significado é o mesmo (`CardGroup`, `AccordionGroup`, `Check`, `cols="3"`, `open="true"`, `logo: ./x.svg`, conteúdo em `<Button>`). Quando não há equivalente, o diagnóstico explica o que usar (`attrAliases`, `valueAliases`, `knownAliases`, `synonyms` em config). |
| Desenhado à mão | Ícones são geometria (`internal/sketch`) passada por uma "caneta" determinística: vértices deslocados, arestas levemente curvas, fechamento que passa do ponto e preenchimento oat fora de registro. Nunca usar arte da Anthropic. Todo traço tem `pathLength="1"` para animar. |
| Movimento | Calmo, como anthropic.com/claude.com: out-quart/out-expo, distâncias pequenas, nada em loop. Hero sobe palavra por palavra com blur; traços se desenham; blocos abaixo da dobra aparecem ao rolar (só o que está fora da tela fica pendente, então nada pisca); abas com indicador deslizante; view transitions entre páginas e na troca de tema. Sem JS ou com `prefers-reduced-motion`, tudo aparece pronto. |
| JS no cliente | Vanilla, mínimo, embutido no binário: abas, copiar código, troca de tema, revelação ao rolar, live reload no `dev`. O site continua completo sem JS. |
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
internal/markdown/   goldmark + parsers de componentes (bloco e inline) + renderers (código, títulos, links)
internal/components/ schemas, aliases de compatibilidade e renderização de cada componente
internal/sketch/     ícones e traços desenhados à mão, gerados por código
internal/site/       varredura da pasta, rotas, navegação, índice global, validação de links
internal/server/     HTTP, snapshot, watcher, last-known-good, live reload
internal/theme/      templates, CSS, JS e fontes embutidos (embed.FS)
site/                site de exemplo: a documentação do próprio Wavemark, escrita em Wavemark
design/              referência visual medida
tools/visual/        Playwright: compare (estilos vs. referência), shots, interact, motion
tools/iconsheet/     folha de prova dos ícones (go run ./tools/iconsheet sheet.html)
```

Dependências: `github.com/yuin/goldmark`, `github.com/alecthomas/chroma/v2` (realce no servidor, por classes CSS), `github.com/fsnotify/fsnotify`, `gopkg.in/yaml.v3`. Evitar dependências novas sem motivo.

## Comandos

```bash
go build -o bin/wavemark ./cmd/wavemark
./bin/wavemark dev site          # servidor de desenvolvimento em :4321
./bin/wavemark check site --json # validação
./bin/wavemark spec              # referência para agentes
go vet ./...

go test ./...

# Verificação no navegador (com o dev rodando em :4321)
cd tools/visual && npm install && npx playwright install chromium
node compare.mjs --scheme light   # estilos computados vs. platform.claude.com e claude.com
node compare.mjs --scheme dark
node interact.mjs                 # abas, copiar, tema, sumário, menu mobile, sem JS, CSP
node motion.mjs                   # quadros da animação, reduced motion, sem JS, indicador das abas
node shots.mjs --pages /,/guia/instalacao --scheme light,dark --scroll 0,900   # capturas em tools/visual/shots/
```

## Teste com agentes às cegas

O teste mais valioso: um subagente sem contexto recebe só o binário (`help`, `spec`, `check --json`), escreve um site completo e relata os erros da primeira execução do `check`. Copie o binário para `bin/wavemark-agent` antes, para que mudanças em andamento não afetem o teste, e use pastas em `.e2e-agents/` (ignoradas pelo git). Rodada 1: 8 e 71 erros na primeira execução. Rodada 2, depois das correções: 0 erros.

## Estado atual

- Implementado: parsers de componentes (bloco e inline), validação com diagnósticos, cascata de config, datas e autores, rotas, navegação (ordem manual ou por data), links, âncoras e pastas validados, layouts `docs`, `article` e `landing` (claro e escuro), ícones desenhados à mão, movimento, runtime com snapshot atômico, watcher, SSE, última versão válida, `llms.txt`, fonte `.md` de cada página, CSP, testes Go.
- `compare.mjs`: 53/53 propriedades iguais à referência em claro e escuro (fontes excluídas por serem proprietárias).
- Desempenho: 601 páginas pesadas reconstroem em ~0,5 s em 2 núcleos. O realce de código é memoizado por hash e a sidebar é renderizada uma vez por snapshot.
- Próximo pedido dos agentes: `spec`, `help` e diagnósticos em inglês (hoje só português), conforme o `lang` do site ou uma flag.
- Ainda não existe: busca, Mermaid, matemática, build estático, i18n da interface além de en/pt, tags de posts.

## Armadilhas conhecidas

- O goldmark chama `Continue` de fora para dentro em cada linha. Por isso a tag de fechamento é decidida em `componentParser.owns`, que olha `pc.OpenedBlocks()`. Mexa ali com cuidado e teste aninhamento, cercas de código com `</Tag>` dentro e componentes autofechados.
- O renderer de tabela do goldmark usa `style=""` por padrão. Mantenha `TableCellAlignAttribute`, porque a CSP bloqueia estilos inline.
- Não use `style=` em nenhum HTML gerado, pelo mesmo motivo; use classes (`wm-columns--3`).
- Ao matar servidores em scripts, use `pgrep -f "^./bin/wavemark"`. Com `pkill -f "wavemark dev"`, o padrão casa com o próprio shell.
- Animações do traço dependem de `pathLength="1"` e `stroke-dasharray: 1 2`. Não use `vector-effect: non-scaling-stroke` nesses paths: ele invalida o `pathLength` e o traço aparece pronto.
- A revelação ao rolar só marca como `pending` o que está fora da tela no carregamento. Nunca esconda conteúdo com CSS sem `.wm-js` e sem respeitar `prefers-reduced-motion`.
- Blocos de código indentados existem no parser só para serem reportados (`markdown.indented-code`). Removê-los do parser fazia o goldmark descartar silenciosamente o resto da página.

## Convenções

- Go idiomático, `gofmt`, erros com contexto (`fmt.Errorf("...: %w", err)`).
- Saída HTML sempre escapada; atributos de URL validados (só http(s), mailto, relativo, `/` e `#`).
- Arquivos servidos só a partir do índice da varredura, nunca por junção de caminho da requisição. Symlinks não são seguidos.
- Documentação e mensagens para o usuário em português; identificadores e código em inglês.
