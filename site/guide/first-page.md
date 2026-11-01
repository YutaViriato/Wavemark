# Your first page

Create a `.md` file anywhere in the folder. The first `#` heading becomes the page title and the rest is plain Markdown.

```markdown title="guide/hello.md"
# Hello

This is my page. It links to [the installation guide](./install.md).

## A section

Some text in the section.
```

The page is served at `/guide/hello` and appears in the sidebar, inside the Guide section.

## Frontmatter

To set the title, description and position in the navigation, add a YAML block at the top:

```yaml title="Frontmatter"
---
title: Hello
description: One sentence shown below the title.
order: 2
---
```

When the frontmatter has a `title`, the body cannot have a `# H1`. The rule keeps one title per page.

## Components

Components are tags with a capital first letter, each on its own line. The content between the tags is Markdown:

```mdx title="A callout"
<Note title="Reminder">
Run `wavemark check` before publishing.
</Note>
```

<Note title="Reminder">
Run `wavemark check` before publishing.
</Note>

The full list is in the [component reference](../reference/components.md).

## Links and anchors

Point links at files, not URLs. Wavemark turns the path into the route and tells you when the target does not exist:

- `[Install](./install.md)` becomes `/guide/install`.
- `[Tabs](../reference/components.md#tabs)` points to a section of another page.
- Anchors come from headings: `## Advanced Setup` becomes `#advanced-setup`.
