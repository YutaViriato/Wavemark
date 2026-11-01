---
title: What blind agents taught us
description: We gave agents only the binary and asked for a whole site. Their first mistakes rewrote the syntax.
date: 2026-11-03
author: The Wavemark team
icon: users
---

The best usability test for a tool written for agents is an agent that has never seen it. We gave three agents nothing but the `wavemark` binary, with `help`, `spec` and `check --json`, and asked each one for a complete website. We kept the raw output of their first `check`.

## Round one

<Columns cols={2}>
<Stat value="8" label="errors on the first check of a developer docs site" />
<Stat value="71" label="errors on the first check of a research studio site" />
</Columns>

The errors were not random. They were habits from other tools, repeated across pages:

| Habit | Where it comes from | Times |
|---|---|---|
| `icon` on `<Card>` | Mintlify | 16 |
| `date` and `author` in the frontmatter | Hugo, Jekyll, Docusaurus | 12 |
| Text between `<Button>` tags | React, MDX | 8 |
| `title` instead of `label` in the navigation | Mintlify, VitePress | 4 |
| `<AccordionGroup>` | Mintlify | 1 |

Both agents also found a real bug: after a code block indented by four spaces, the rest of the page silently disappeared, links and headings included. `check` said everything was fine.

## What changed

We took each habit and asked one question: does it mean the same thing in Wavemark? When it did, Wavemark started accepting it. When it did not, the diagnostic started naming what to use instead.

- `icon`, `date`, `author` and button content became real features.
- `<CardGroup>`, `<AccordionGroup>`, `<Check>`, `cols="3"` and `logo: ./logo.svg` became aliases.
- Icon names from Lucide and Font Awesome map to the hand-drawn set.
- Indented code became an explicit error, and the lost content came back.

## Round two

<Stat value="0" label="errors on the first check, with the fixed binary" />

The third agent built a full site for a fictional research lab: a landing page, four dated posts, documentation with version badges and an FAQ. Its first `check` had no errors. Its one request was the reason this site is now in English: the reference and the diagnostics were only in Portuguese.

<Quote author="The third agent" role="Final report">
The hints are among the best I've seen. The one change that would help most is English.
</Quote>
