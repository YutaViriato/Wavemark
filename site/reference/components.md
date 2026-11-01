---
title: Components
description: Every component, with the source and the result side by side.
---

Components are PascalCase tags. Each tag sits on its own line and the content between the tags is Markdown. The reference generated from the schemas is printed by `wavemark spec`.

## Callouts

Use `<Callout>` with a `type`, or one of the shorthands: `<Note>`, `<Tip>`, `<Info>`, `<Success>` (or `<Check>`), `<Warning>` and `<Danger>`.

```mdx
<Tip title="Shorthand">
`<Tip>` is the same as `<Callout type="tip">`.
</Tip>
```

<Note>
A neutral note without a title, for remarks that do not change the outcome.
</Note>

<Tip title="Shorthand">
`<Tip>` is the same as `<Callout type="tip">`.
</Tip>

<Info title="Good to know">
Use it for context that helps but that readers can skip.
</Info>

<Success title="Done">
Confirms that something worked.
</Success>

<Warning title="Careful">
Something can go wrong if the reader ignores this.
</Warning>

<Danger title="Danger">
The action cannot be undone.
</Danger>

## Cards and columns

`<Columns>` makes a grid of 1 to 4 columns (`<CardGroup>` is a shorthand with the same effect). On narrow screens it becomes a single column. A `<Card>` with `href` is a link as a whole, and `icon` adds a [hand-drawn icon](./icons.md).

```mdx
<Columns cols={2}>
<Card title="Installation" icon="rocket" href="../guide/install.md" cta="Install">
Build the binary and serve a folder.
</Card>
<Card title="Configuration" icon="gear" href="../guide/configuration.md" cta="Configure">
Three levels of configuration with inheritance.
</Card>
</Columns>
```

<Columns cols={2}>
<Card title="Installation" icon="rocket" href="../guide/install.md" cta="Install">
Build the binary and serve a folder.
</Card>
<Card title="Configuration" icon="gear" href="../guide/configuration.md" cta="Configure">
Three levels of configuration with inheritance.
</Card>
</Columns>

<Columns cols={3}>
<Card title="No link">
A card without `href` is just a block of content.
</Card>
<Card title="Markdown inside">
- Lists
- **Emphasis**
- `code`
</Card>
<Card title="Equal height">
Cards in the same row share the same height.
</Card>
</Columns>

## Tabs

```mdx
<Tabs>
<Tab title="macOS">
Install with Homebrew.
</Tab>
<Tab title="Linux">
Download the binary.
</Tab>
</Tabs>
```

<Tabs>
<Tab title="macOS">
Install with Homebrew and check the version:

```bash
brew install wavemark
wavemark help
```
</Tab>
<Tab title="Linux">
Download the binary and put it on your `PATH`.
</Tab>
<Tab title="Windows">
Use `wavemark.exe`.
</Tab>
</Tabs>

## Code groups

`<CodeGroup>` accepts only fenced code blocks. Each tab label comes from `title="..."` on the fence or, without it, from the language name.

<CodeGroup>
```python title="main.py"
import anthropic

client = anthropic.Anthropic()
message = client.messages.create(
    model="claude-opus-5-5",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello, Claude"}],
)
print(message.content)
```
```typescript title="main.ts"
import Anthropic from "@anthropic-ai/sdk";

const client = new Anthropic();
const message = await client.messages.create({
  model: "claude-opus-5-5",
  max_tokens: 1024,
  messages: [{ role: "user", content: "Hello, Claude" }],
});
console.log(message.content);
```
```go title="main.go"
package main

import "fmt"

func main() {
	fmt.Println("Hello, Claude")
}
```
</CodeGroup>

## Steps

<Steps>
<Step title="Write">
Create a `.md` file in the site folder.
</Step>
<Step title="Check">
Run `wavemark check` and fix the errors.
</Step>
<Step title="Publish">Run `wavemark serve`.</Step>
</Steps>

## Flow diagrams

`<Flow>` draws steps connected by hand-drawn arrows. On narrow screens it turns vertical.

```mdx
<Flow caption="From file to response.">
<FlowStep title="Markdown" icon="folder">your files</FlowStep>
<FlowStep title="Parser" icon="code">validated</FlowStep>
<FlowStep title="HTTP" icon="globe">served</FlowStep>
</Flow>
```

<Flow caption="From file to response.">
<FlowStep title="Markdown" icon="folder">your files</FlowStep>
<FlowStep title="Parser" icon="code">validated</FlowStep>
<FlowStep title="HTTP" icon="globe">served</FlowStep>
</Flow>

## Collapsible sections

<Accordion title="Can I use HTML?">
No. Raw HTML is rejected so the theme stays consistent. If a component is missing, it should become part of Wavemark.
</Accordion>

<Accordion title="What about JavaScript in components?">
No either. Attribute values are literals: strings, numbers, booleans or JSON.
</Accordion>

For a list of questions, `<AccordionGroup>` joins the items into one block:

<AccordionGroup>
<Accordion title="Does Wavemark build a static site?">
Not yet. It is a runtime: it serves the folder directly.
</Accordion>
<Accordion title="Does it work without JavaScript in the browser?">
Yes. Tabs show all of their content, and everything else works as usual.
</Accordion>
<Accordion title="Does it respect reduced motion?">
Yes. When the system preference is on, no animation runs.
</Accordion>
</AccordionGroup>

## Buttons

Consecutive buttons sit side by side. The text comes from `label` or from the content.

```mdx
<Button href="../guide/install.md" label="Get started" />
<Button href="https://github.com/YutaViriato/Wavemark" variant="secondary">Source code</Button>
```

<Button href="../guide/install.md" label="Get started" />
<Button href="https://github.com/YutaViriato/Wavemark" variant="secondary">Source code</Button>

## Quotes

<Quote author="Ada Lovelace" role="Notes on the Analytical Engine, 1843">
The Analytical Engine has no pretensions whatever to originate anything. It can do whatever we know how to order it to perform.
</Quote>

## Numbers

<Columns cols={3}>
<Stat value="15" label="pages in this site" />
<Stat value="52" label="hand-drawn icons" />
<Stat value="0" label="lines of raw HTML" />
</Columns>

## Badges and inline icons

`<Badge>` and `<Icon />` are inline: they go inside paragraphs, headings and tables, and open and close on the same line.

```mdx
Static export <Badge color="clay">Soon</Badge> and search <Badge>Planned</Badge>.
Made with <Icon name="heart" label="love" /> and <Icon name="coffee" label="coffee" />.
```

Static export <Badge color="clay">Soon</Badge> and search <Badge>Planned</Badge>. Colors: <Badge>gray</Badge> <Badge color="clay">clay</Badge> <Badge color="olive">olive</Badge> <Badge color="sky">sky</Badge> <Badge color="fig">fig</Badge>.

Made with <Icon name="heart" label="love" /> and <Icon name="coffee" label="coffee" />.

## Page lists

`<PageList />` lists the pages of a folder, with date, author and description. With `sort="auto"` (the default), dated pages go from newest to oldest.

```mdx
<PageList dir="../blog" />
```

<PageList dir="../blog" />

## Landing page only

`<Hero>` opens the page with a large title that arrives word by word. `highlight` underlines a phrase with a hand-drawn stroke; `image` puts an image beside the text. `<Section tone="muted|oat|clay|dark">` makes full-width bands. See the [home page](../index.md).

```mdx
<Hero title="Markdown that becomes a site" highlight="becomes a site" eyebrow="Wavemark">
A subtitle in Markdown.

<Button href="/guide" label="Get started" />
</Hero>

<Section tone="dark">
<Quote author="Someone">One sentence.</Quote>
</Section>
```
