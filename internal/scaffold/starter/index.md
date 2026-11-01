---
title: {{TITLE}}
description: A site made with Wavemark.
layout: landing
---

<Hero eyebrow="{{TITLE}}" title="A new site, ready to write" highlight="ready to write">
Every `.md` file in this folder is a page. Edit this one, save, and watch the browser update.

<Button href="./getting-started.md" label="Getting started" />
</Hero>

## What's here

<Columns cols={3}>
<Card title="Pages" icon="file" href="./getting-started.md" cta="Open">
`getting-started.md` is a documentation page with a sidebar and a table of contents.
</Card>
<Card title="Components" icon="layers">
Cards, tabs, steps, callouts and more. Run `wavemark spec` to see all of them.
</Card>
<Card title="Checks" icon="shield">
Run `wavemark check` to catch broken links, unknown attributes and other mistakes.
</Card>
</Columns>
