# Wavemark

Wavemark is a runtime that turns a folder of Markdown files into a website. You point the binary at a folder and the server renders the pages; editing a `.md` file changes the site with no build step.

Content is mostly written by AI agents. That drives almost every decision below: a strict grammar, schema validation, actionable errors and a machine-readable reference.

**Everything in this repository is in English**: code, messages, docs, the example site and commit messages. The only other-language text is data for sites written in other languages (interface strings in `internal/i18n` and month names in `internal/components/date.go`).

**The Wavemark site itself is made with Wavemark.** Every page in `site/` is Markdown and components; nothing is hand-made HTML, CSS or images. If a page needs something Wavemark cannot express, add it to Wavemark (a component, an icon, a layout) and then use it.

## Product decisions

| Topic | Decision |
|---|---|
| Execution model | A real runtime: an HTTP server reads the folder, keeps a snapshot in memory and rebuilds when files change (fsnotify). No static build for now. |
| Audience | Documentation and long-form notes, plus site components (cards, columns, hero, buttons, sections) for entry pages and blogs. |
| Component syntax | Declarative JSX-style tags, **no JavaScript**: `<Card title="X" href="/y">`. A capitalized name is a component. Attribute values: `"string"`, `'string'`, `{JSON literal}` or a bare attribute for true. JS expressions are errors. |
| Grammar rules | Block component tags sit on their own line (opening, closing, self-closing). Exception: the one-line form `<Step title="x">text</Step>`. Opening tags may span lines. Content between tags is Markdown and may be indented (the component strips the indentation). Inline components (`<Badge>`, `<Icon />`) live inside text and open and close on the same line. |
| Raw HTML | Forbidden. `check` reports it and it is not rendered, so agents cannot break the visual identity. |
| Indented code blocks | Reported as errors (`markdown.indented-code`); only fenced code is allowed. They stay in the parser only so the content after them is kept and checked. |
| Configuration | Three-level cascade: `wavemark.yaml` (root, `defaults` key) → `_meta.yaml` per folder (`defaults` key) → page frontmatter. Maps merge deeply, scalars override, lists replace. Unknown keys are errors (catches `tittle:`). Site keys (`nav`, site `title`, `logo`, `footer`, `fonts`) only exist in `wavemark.yaml`. |
| Page title | The frontmatter `title` or, without it, the first `# H1` (which leaves the body and goes to the layout). An H1 in the body plus a frontmatter `title` is an error. More than one H1 is an error. |
| Routes | From folders: `guide/install.md` → `/guide/install`; `index.md` or `README.md` → the folder route. Each page's source is also served at `<route>.md`, and the site exposes `/llms.txt`. A root `404.md` becomes the not-found page. |
| Internal links | Relative `.md` links are rewritten to routes and validated against the global index, anchors included. A broken link is an error. |
| Invalid pages | `wavemark serve` only starts if `check` passes. While running, a page that becomes invalid keeps serving its last valid version and the error goes to the log. `wavemark dev` overlays errors on the page and reloads via SSE; pages also reload when the server restarts (boot id). |
| Layouts | `docs` (default: sidebar, content, table of contents on the right, neutral `#fcfcfb` background like platform.claude.com/docs), `article` (long-form reading like anthropic.com/research: numbered chapters on the left, facts on the right, 625 px serif body) and `landing` (hero, large titles, ivory `#faf9f5` background like claude.com). |
| Components | Block: `Callout` (+ `Note`, `Tip`, `Info`, `Success`/`Check`, `Warning`, `Danger`), `Card` (with `icon`), `Columns` (+ `CardGroup`), `Tabs`/`Tab`, `CodeGroup`, `Steps`/`Step`, `Flow`/`FlowStep`, `Accordion`, `AccordionGroup`, `Button`, `Hero` (with `highlight` and `image`), `Section`, `Quote`, `Stat`, `PageList`, `Figure`. Inline: `Badge`, `Icon`. Mermaid, math and search come later. |
| Compatibility | Names agents bring from Mintlify/Docusaurus become aliases when they mean the same thing (`CardGroup`, `AccordionGroup`, `Check`, `cols="3"`, `open="true"`, `logo: ./x.svg`, content inside `<Button>`). When there is no equivalent, the diagnostic says what to use (`attrAliases`, `valueAliases`, `knownAliases`, config `synonyms`). |
| Hand-drawn look | Icons and arrows are geometry (`internal/sketch`) traced by a deterministic "pen": nudged corners, slightly bowed edges, a close that overshoots, and an oat fill out of register. Never use Anthropic artwork. Every stroke has `pathLength="1"` so it can animate. |
| Motion | Calm, like anthropic.com and claude.com: out-quart/out-expo, small distances, nothing loops. The hero rises word by word with blur; strokes draw themselves; blocks below the fold appear on scroll (only offscreen blocks become pending, so nothing flickers); tabs have a sliding indicator; view transitions between pages and on theme change. Without JS or with `prefers-reduced-motion`, everything is simply there. |
| Client JS | Vanilla and minimal, embedded in the binary: tabs, copy code, theme toggle, scroll reveal, live reload in `dev`. The site is complete without JS. |
| Distribution | Out of scope for now. |

## Visual identity

The reference is the current sites (measured November 2026), not the older public guidelines (Poppins/Lora). Measured values and method are in `design/reference.md`.

- Anthropic Sans/Serif/Mono are proprietary. **Never embed or hotlink them.** We use open-licensed alternatives chosen by visual comparison: Hanken Grotesk (sans), Newsreader (serif) and JetBrains Mono. The config lets a site point at font files it is licensed for.
- Do not ship the Claude logo or the Anthropic wordmark. The logo comes from the site config.
- Every visual value comes from CSS tokens (`--wm-*`). Components never use literal colors.
- Visual verification: `tools/visual/` (Playwright) captures Wavemark pages and the reference pages at the same viewport and compares computed styles. Pixel diffs do not work because the fonts differ.

## For agents writing content

- `wavemark spec` prints the full component and frontmatter reference, generated from the same schemas that validate. Read it first.
- `wavemark check [dir] --json` returns diagnostics with file, line, column, code and hint. Run it until it is clean.

## Architecture (Go)

```
cmd/wavemark/        CLI: init, dev, browse, check, build, serve, spec, version
internal/diag/       diagnostics (file, line, column, code, message, hint)
internal/config/     wavemark.yaml, _meta.yaml, frontmatter, cascade and key validation
internal/markdown/   goldmark + component parsers (block and inline) + renderers (code, headings, links)
internal/components/ schemas, compatibility aliases and rendering of each component
internal/sketch/     hand-drawn icons, arrows and strokes, generated by code
internal/site/       folder scan, routes, navigation, global index, link validation, llms.txt
internal/server/     HTTP, snapshot, watcher, last-known-good, live reload (pluggable builder for browse)
internal/browse/     'wavemark browse': any project's Markdown, lenient (ParsePlain), listings per folder
internal/export/     static export ('wavemark build')
internal/scaffold/   starter site ('wavemark init')
internal/theme/      embedded templates, CSS, JS and fonts (embed.FS)
site/                the example site: Wavemark's own documentation, written in Wavemark
design/              measured visual reference
tools/visual/        Playwright: compare (styles vs. reference), shots, interact, motion
tools/iconsheet/     icon contact sheet (go run ./tools/iconsheet sheet.html)
tools/showreel/      the 30 s intro video and its renderer: real pages, icons and a synthesized score
```

Dependencies: `github.com/yuin/goldmark`, `github.com/alecthomas/chroma/v2` (server-side highlighting with CSS classes), `github.com/fsnotify/fsnotify`, `gopkg.in/yaml.v3`. Avoid new dependencies without a reason.

## Commands

```bash
go build -o bin/wavemark ./cmd/wavemark
./bin/wavemark dev site          # development server on :4321
./bin/wavemark check site --json # validation
./bin/wavemark spec              # reference for agents
go vet ./...
go test ./...

# Browser checks (with dev running on :4321)
cd tools/visual && npm install && npx playwright install chromium
node compare.mjs --scheme light   # computed styles vs. platform.claude.com and claude.com
node compare.mjs --scheme dark
node interact.mjs                 # tabs, copy, theme, table of contents, mobile menu, no-JS, CSP
node motion.mjs                   # animation frames, reduced motion, no-JS, tab indicator
node shots.mjs --pages /,/guide/install --scheme light,dark --scroll 0,900   # screenshots in tools/visual/shots/
```

## Blind agent tests

The most useful test: a subagent with no context gets only the binary (`help`, `spec`, `check --json`), writes a complete site and reports the errors of its first `check` run. Copy the binary to `bin/wavemark-agent` first so work in progress does not affect the test, and use folders in `.e2e-agents/` (ignored by git). Round 1: 8 and 71 errors on the first run. Round 2, after the fixes: 0 errors.

## Current state

- Done: block and inline component parsers, validation with diagnostics, config cascade, dates and authors, routes, navigation (manual or by date), validated links, anchors and folders, `docs`, `article` and `landing` layouts (light and dark), hand-drawn icons and flow diagrams, motion, a runtime with an atomic snapshot, watcher, SSE, last valid version, `llms.txt`, Markdown source for each page, Markdown 404 page, CSP, Go tests.
- `compare.mjs`: 53/53 properties match the reference in light and dark (fonts excluded because they are proprietary).
- Performance: 601 heavy pages rebuild in ~0.5 s on 2 cores. Highlighting is memoized by hash and the sidebar is rendered once per snapshot.
- Not yet: search, Mermaid, math, static export, interface strings beyond en/pt, post tags.

## wavemark browse

`wavemark browse [folder]` is the one place where Markdown is read leniently: files from any project, not written for Wavemark. `markdown.ParsePlain` uses goldmark GFM without component parsers or `{#id}` attributes; `RenderOptions.Plain` renders indented code, GitHub alerts as callouts and raw HTML through `htmlSanitizer` (rebuilds every tag from an allow-list, keeps only `href`/`src`/`alt`/`title`/`width`/`height`/`align`/`open` with checked URLs, keeps tags balanced, shows anything else as escaped text). Routes are the file paths (`/docs/plan.md`), folder listings are `/docs`, raw sources are `/_wavemark/source/<path>`. Scan uses `git ls-files --cached --others --exclude-standard` when possible, always skips dependency folders and symlinks, and only serves Markdown and images. It binds to 127.0.0.1 by default. `Site.Version` lets the server skip reloads when no Markdown or image changed.

## Known pitfalls

- goldmark calls `Continue` outermost to innermost on every line. That is why the closing tag is decided in `componentParser.owns`, which inspects `pc.OpenedBlocks()`. Change it carefully and test nesting, fences containing `</Tag>`, and self-closing components.
- goldmark's table renderer uses `style=""` by default. Keep `TableCellAlignAttribute`, because the CSP blocks inline styles.
- Never emit `style=` in generated HTML, for the same reason; use classes (`wm-columns--3`).
- When killing servers in scripts, use `pgrep -f "^./bin/wavemark"`. With `pkill -f "wavemark dev"` the pattern matches your own shell.
- Stroke animations depend on `pathLength="1"` and `stroke-dasharray: 1 2`. Do not use `vector-effect: non-scaling-stroke` on those paths: it breaks `pathLength` and the stroke shows up already drawn.
- Scroll reveal only marks blocks that are offscreen at load as `pending`. Never hide content with CSS without `.wm-js` and without respecting `prefers-reduced-motion`.

## Conventions

- Idiomatic Go, `gofmt`, errors with context (`fmt.Errorf("...: %w", err)`).
- HTML output is always escaped; URL attributes are validated (only http(s), mailto, tel, relative, `/` and `#`).
- Files are served only from the scan index, never by joining the request path. Symlinks are not followed.
- English everywhere: identifiers, comments, messages, docs, site content and commits.
