---
title: CLI
description: The commands of the wavemark binary.
---

## wavemark init

```bash
wavemark init [folder]
```

Creates a starter site: `wavemark.yaml`, a landing page and a documentation page. It never overwrites files; if the folder already has a site, it stops.

## wavemark dev

```bash
wavemark dev [folder] [--addr :4321]
```

Serves the folder, reloads the browser on every change and shows each page's errors in an overlay. The terminal lists every diagnostic after each rebuild. When the server restarts with a new binary, open pages reload by themselves.

## wavemark browse

```bash
wavemark browse [folder] [--no-open] [--all] [--addr localhost:4400]
```

Opens every Markdown file of a project in the browser: READMEs, `CLAUDE.md` and `AGENTS.md`, plans, specs and notes that agents write. It works in any folder; the files do not need to be a Wavemark site.

- **The listing** shows the files that changed most recently, then every file by folder, with its title, first paragraph and age. Type `/` to filter by name, path or text. Each folder has its own listing.
- **Each file** is rendered with the Wavemark theme: sidebar with the file tree, table of contents, highlighted code, frontmatter as a table, and GitHub alerts (`> [!NOTE]`) as callouts. Relative links between files work.
- **It stays current.** New and edited files appear while the page is open. Saving a file that is not Markdown does not reload anything.
- **Agent files are labelled:** instructions (`CLAUDE.md`, `AGENTS.md`, `.kiro/steering/`, `.cursor/rules/`, `.github/copilot-instructions.md`…), specs (`.kiro/specs/`), skills, commands and subagents.

The files are read as plain GitHub-flavoured Markdown, so nothing is validated. Raw HTML is cleaned: simple tags such as `<details>`, `<kbd>`, `<br>` and `<img>` render, and anything else shows as text. Indented code renders as code.

Inside a git repository, files ignored by `.gitignore` are left out; `--all` includes them. Dependency and cache folders (`node_modules`, `vendor`, `.venv`…) are always skipped, and only Markdown and images are served.

The server listens on `localhost` only, from port 4400 up to the first free one, and opens your browser (`$BROWSER` if set). `--no-open` only prints the address. With `--addr`, a host such as `0.0.0.0:4400` exposes the files to your network; the command warns when it does.

## wavemark build

```bash
wavemark build [folder] [--out dist] [--base /path]
```

Writes the site as static files: one `.html` per page, each page's Markdown source, `404.html`, `llms.txt` and the theme assets. It refuses to write anything when there are errors. `--base` sets the URL path the site is published under; by default it is the path of `baseURL` in `wavemark.yaml`. The output folder is replaced on each build, but only if a previous build created it.

## wavemark serve

```bash
wavemark serve [folder] [--addr :8080]
```

The production server. It refuses to start when there are errors. Once it is up, a page that starts having errors keeps its last valid version.

Theme assets are served with a hash in their path and an immutable cache. Pages use `ETag`. A `404.md` at the root, if present, is the not-found page.

## wavemark check

```bash
wavemark check [folder] [--json]
```

Validates everything and exits with status 1 when there are errors. With `--json`, the output is:

```json
{
  "ok": false,
  "pages": 9,
  "errors": 1,
  "warnings": 0,
  "diagnostics": [
    {
      "file": "guide/x.md",
      "line": 12,
      "column": 5,
      "severity": "error",
      "code": "component.attr-unknown",
      "message": "unknown attribute 'titl' on <Card>",
      "hint": "Did you mean 'title'? Valid attributes: title, href, icon, cta."
    }
  ]
}
```

The `code` values are stable and safe for tools to depend on.

## wavemark spec

```bash
wavemark spec [--json]
```

Prints the reference of the syntax, components and configuration. It is generated from the same schemas that validate content. With `--json`, it prints the component schemas, the icon names and the layouts.

## wavemark version

```bash
wavemark version
```

Prints the installed version. To upgrade, run the install script again.
