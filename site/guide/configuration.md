---
title: Configuration
description: Three levels of configuration, from the whole site down to one page, with inheritance.
---

Configuration has three levels. Each more specific level overrides the one before it:

1. `wavemark.yaml`, at the root, applies to the whole site.
2. `_meta.yaml`, in each folder, applies to the folder and its subfolders.
3. The frontmatter applies only to its page.

## wavemark.yaml

```yaml title="wavemark.yaml"
title: Acme Docs
description: Documentation for Acme.
lang: en
nav:
  - label: Guide
    href: /guide
footer:
  text: © 2026 Acme
defaults:
  toc: true
```

The keys `title`, `nav`, `logo`, `footer` and `fonts` only exist here. The `defaults` key sets what every page inherits.

## _meta.yaml

```yaml title="guide/_meta.yaml"
title: Guide
order: [index, install, first-page]
defaults:
  layout: docs
```

`order` lists file names without `.md` and subfolder names. Anything not listed comes after, sorted by the frontmatter `order` and then by title. With `sort: date`, unlisted pages come newest first, which is what a blog wants.

## The cascade

<AccordionGroup>
<Accordion title="An inheritance example" open>
With `toc: false` in `wavemark.yaml` and `toc: true` in `guide/_meta.yaml`, every page in `guide/` shows the table of contents and the others do not. A page inside `guide/` can still write `toc: false` in its frontmatter.
</Accordion>
<Accordion title="What is not inherited">
`title`, `description`, `order`, `date` and `author` belong to the page. Only the `defaults` keys (`layout`, `toc` and `hidden`) pass from one level to the next.
</Accordion>
<Accordion title="Defaults reach the folder's own index">
A folder's `defaults` also apply to its `index.md`. A blog folder with `defaults: {layout: article}` turns its listing page into an article too, unless the listing page sets `layout: docs` itself.
</Accordion>
</AccordionGroup>

## Your own fonts

The theme uses open-licensed fonts by default. If you have a license for other fonts, point at the `.woff2` files in `wavemark.yaml`:

```yaml
fonts:
  sans: fonts/my-sans.woff2
  serif: fonts/my-serif.woff2
```

<Danger title="Font licenses">
Do not use Anthropic's fonts without a license. They are proprietary.
</Danger>
