---
title: Overview
description: What Wavemark does, how it maps a folder to a site, and what happens when a file changes.
---

Wavemark is a runtime: a process that serves a folder of Markdown as a website. There is no build step and no `dist/` folder. You edit a file, the server notices the change and the next request already gets the new page.

## How a folder becomes a site

Every `.md` file is a page and every folder is a section of the navigation. The route comes from the file path:

| File | Route |
|---|---|
| `index.md` | `/` |
| `guide/index.md` | `/guide` |
| `guide/install.md` | `/guide/install` |
| `reference/components.md` | `/reference/components` |

Images and other files are served at the same path they have in the folder.

## What happens when a file changes

<Flow>
<FlowStep title="You save" icon="pencil">a file in the folder</FlowStep>
<FlowStep title="Debounce" icon="clock">120 ms to group saves</FlowStep>
<FlowStep title="Rebuild" icon="gear">parse, validate, render</FlowStep>
<FlowStep title="Swap" icon="bolt">one atomic switch</FlowStep>
</Flow>

The server keeps an immutable *snapshot* of the site in memory: rendered pages, navigation and the link index. When the file system reports a change, Wavemark waits 120 ms to group saves in a row, rebuilds the snapshot and swaps the old one for the new one at once. A request in flight never sees half a site.

<Callout type="tip" title="Last valid version">
In production, if a page starts having errors, the server keeps serving its last valid version and logs the problem. The other pages keep updating.
</Callout>

## Why so strict

Wavemark content is often written by AI agents. An agent cannot see the rendered page, so every mistake has to show up as text: the file, the line, the column, a code and a fix. That is why:

- Raw HTML is not allowed. The visual identity comes only from components.
- Internal links are validated, anchors included.
- Unknown configuration keys are errors, not silence.

See [For agents](../agents.md) for the whole write-and-check loop.
