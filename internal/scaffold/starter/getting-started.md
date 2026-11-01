---
title: Getting started
description: How to work on this site.
---

This page uses the `docs` layout: navigation on the left, the table of contents on the right.

## Preview

```bash
wavemark dev .
```

Open the address it prints. Every save reloads the page.

## Write

<Steps>
<Step title="Add a page">
Create a `.md` file. Its path becomes its address: `guide/setup.md` is served at `/guide/setup`.
</Step>
<Step title="Use components">
Run `wavemark spec` for the full reference of Markdown rules and components.
</Step>
<Step title="Check">
Run `wavemark check .` and fix what it reports.
</Step>
</Steps>

## Publish

<Tabs>
<Tab title="Static files">
`wavemark build . --out ../dist` writes plain HTML for GitHub Pages, Netlify or any web server.
</Tab>
<Tab title="Server">
`wavemark serve .` runs the site as a live server, usually behind a reverse proxy.
</Tab>
</Tabs>

<Tip>
Links between pages point to files: `[Setup](./guide/setup.md)`. Wavemark turns them into addresses and reports the ones that do not exist.
</Tip>
