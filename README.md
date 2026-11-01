# Wavemark

Wavemark turns a folder of Markdown files into a website, in the visual language of Anthropic and Claude. One binary, no dependencies, no configuration required.

## Install

```bash
curl -fsSL https://yutaviriato.github.io/Wavemark/install.sh | sh
```

On Windows (PowerShell):

```powershell
irm https://yutaviriato.github.io/Wavemark/install.ps1 | iex
```

Or download an archive from [Releases](https://github.com/YutaViriato/Wavemark/releases).

## Use

```bash
wavemark init my-site              # a starter site
wavemark dev my-site               # preview at http://localhost:4321, reloads on save
wavemark check my-site             # broken links, unknown components, raw HTML…
wavemark build my-site --out dist  # static HTML for GitHub Pages or any web server
wavemark serve my-site             # or run it as a live server
```

`wavemark spec` prints the complete syntax and component reference.

## For AI agents

Give your agent one sentence:

> Use Wavemark from https://yutaviriato.github.io/Wavemark/install-wavemark and create a site about X.

That page tells the agent how to install Wavemark, read the reference, write pages and validate them with `wavemark check --json` until there are no errors.

## Documentation

The documentation is a Wavemark site in [`site/`](site/), published at https://yutaviriato.github.io/Wavemark/.

## Development

```bash
go build -o bin/wavemark ./cmd/wavemark
go test ./...
./bin/wavemark dev site
```

See [CLAUDE.md](CLAUDE.md) for the design decisions and architecture.
