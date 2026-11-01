---
title: Wavemark
description: Point it at a folder of Markdown and it becomes a website.
layout: landing
---

<Hero eyebrow="Wavemark" title="Markdown that becomes a site while you write" highlight="while you write">
One binary, a folder of `.md` files and no build step. Made for long-form documentation written by people and by agents.

<Button href="./guide/install.md" label="Get started" />
<Button href="./reference/components.md" label="See the components" variant="secondary" />
</Hero>

## Write Markdown. The rest is runtime.

<Columns cols={3}>
<Card title="No build step" icon="bolt" href="./guide/index.md" cta="How it works">
The server reads the folder, keeps a snapshot in memory and rebuilds it when a file changes.
</Card>
<Card title="Strict on purpose" icon="shield" href="./agents.md" cta="For agents">
An unknown attribute, a broken link or raw HTML becomes an error with a line, a column and a fix.
</Card>
<Card title="Declarative components" icon="layers" href="./reference/components.md" cta="Reference">
Cards, tabs, steps and callouts as JSX-style tags, with no JavaScript in your content.
</Card>
</Columns>

<Columns cols={4}>
<Stat value="1" label="binary, with theme, fonts and scripts embedded" />
<Stat value="0" label="build steps between saving and publishing" />
<Stat value="~0.5 s" label="to rebuild 600 pages" />
<Stat value="29" label="components, each validated by a schema" />
</Columns>

## From file to page

<Flow caption="What happens when you save a file.">
<FlowStep title="Your folder" icon="folder">Markdown, images, YAML</FlowStep>
<FlowStep title="Parser" icon="code">goldmark and schemas</FlowStep>
<FlowStep title="Snapshot" icon="layers">rebuilt in memory</FlowStep>
<FlowStep title="Browser" icon="globe">served, reloaded live</FlowStep>
</Flow>

<CodeGroup>
```bash title="Development"
wavemark dev site
```
```bash title="Validation"
wavemark check site --json
```
```bash title="Production"
wavemark serve site --addr :8080
```
</CodeGroup>

<Section tone="dark">
<Quote author="The Wavemark team" role="Design principle">
If an agent cannot see the page, every mistake has to exist as text: the file, the line, the column and what to do about it.
</Quote>
</Section>

## From the blog

<PageList dir="./blog" style="cards" limit={3} />

<Section tone="oat" align="center">
## Start with a folder

Create a `.md` file, run `wavemark dev` and open your browser.

<Button href="./guide/install.md" label="Install Wavemark" />
</Section>
