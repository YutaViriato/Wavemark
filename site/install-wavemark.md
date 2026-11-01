---
title: Install Wavemark
description: Everything an AI agent needs to install Wavemark and build a site with it.
order: 1
---

Wavemark turns a folder of Markdown files into a website. It is a single binary with no dependencies. If you are an agent asked to build a site with Wavemark, follow these steps in order.

## 1. Install

<Tabs>
<Tab title="Linux and macOS">
```bash
curl -fsSL https://wavemark.viriato.site/install.sh | sh
export PATH="$HOME/.local/bin:$PATH"
wavemark version
```
</Tab>
<Tab title="Windows">
```powershell
irm https://wavemark.viriato.site/install.ps1 | iex
wavemark version
```
</Tab>
</Tabs>

The script downloads the binary for your system from GitHub Releases, verifies its checksum and installs it in `~/.local/bin` (Windows: `%LOCALAPPDATA%\wavemark`). It needs no root access and no other software.

## 2. Read the reference

```bash
wavemark spec
```

Read the whole output before writing. It lists every Markdown rule, component, attribute, icon and configuration key, and it is generated from the same code that validates your files. The rules that matter most:

- Raw HTML is not allowed. Use Markdown and components.
- Block component tags each sit on their own line.
- One title per page: `title` in the frontmatter or a `# H1`, not both.
- Links point to `.md` files: `[Setup](./setup.md)`.

## 3. Create the site

```bash
wavemark init my-site
```

This writes a valid starter site: `wavemark.yaml`, a landing page and a docs page. Edit and add `.md` files. Every file is a page, and every folder is a section of the navigation.

## 4. Check until it is clean

```bash
wavemark check my-site --json
```

Each diagnostic has a file, a line, a column, a code and a hint. Fix them and run the check again until it reports `"ok": true`. Warnings do not block, but fix them too.

## 5. Preview or publish

<Columns cols={2}>
<Card title="Preview" icon="eye">
`wavemark dev my-site` serves the site at `http://localhost:4321` and reloads on every save.
</Card>
<Card title="Publish" icon="rocket">
`wavemark build my-site --out dist` writes static HTML for GitHub Pages, Netlify or any web server.
</Card>
</Columns>

## Reading this page as Markdown

Every page of this site is also available as plain Markdown: add `.md` to its address, as in `/install-wavemark.md`. The whole site is mapped in [llms.txt](/llms.txt).
