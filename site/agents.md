---
title: For agents
description: How an AI agent should write and validate content for Wavemark.
order: 10
---

Wavemark is designed for content written by agents. An agent cannot see the rendered page, so the runtime turns every problem into a text diagnostic.

## The loop

<Steps>
<Step title="Read the reference">
Run `wavemark spec` and put its output in your context. It describes every component, attribute and rule, and it is generated from the code itself.
</Step>
<Step title="Write the files">
Create the `.md` files in the site folder. Point links at files (`./other.md`), never at made-up URLs.
</Step>
<Step title="Check">
Run `wavemark check site --json`. Every diagnostic has a file, a line, a column, a code and, almost always, a hint.
</Step>
<Step title="Fix until it is clean">
Repeat until `"ok": true`. Warnings do not block, but they are worth fixing.
</Step>
</Steps>

## Common mistakes

| Code | Cause | Fix |
|---|---|---|
| `html.forbidden` | `<br>`, `<div>` or other HTML | Use Markdown or a component |
| `component.unknown` | A name from another tool | Follow the hint, which names the equivalent |
| `component.attr-unknown` | `icon` where it does not exist, `subtitle` on `<Hero>` | The hint lists the valid attributes |
| `link.broken` | A link to a file that does not exist | Use the path of the `.md` file, relative to the page |
| `link.anchor` | A `#section` that does not exist | Anchors come from headings; the hint lists the existing ones |
| `title.duplicate` | `title:` in the frontmatter and a `# H1` in the body | Pick one |
| `markdown.indented-code` | A block indented by 4 spaces | Use a fenced block |

## Familiar names work

Agents bring habits from Mintlify, Docusaurus and MDX. When a name means the same thing, Wavemark accepts it: `<CardGroup>`, `<AccordionGroup>`, `<Check>`, `cols="3"`, `open="true"`, text between `<Button>` tags, `logo: ./logo.svg`, and icon names from Lucide and Font Awesome. When there is no equivalent, the diagnostic says what to use instead.

<Info title="Why not MDX">
MDX runs JavaScript and accepts any HTML. That makes content impossible to validate and easy to break visually. Wavemark keeps the look of tags, but every tag has a schema.
</Info>

## Results so far

In blind tests, agents with no prior context built complete sites using only the binary. The first round produced 8 and 71 errors on the first check. After the fixes those runs suggested, the next agent's first check had 0 errors. The full story is on the [blog](./blog/what-blind-agents-taught-us.md).
