---
title: Installation
description: Build the binary and serve your first folder in under a minute.
---

Wavemark is a single binary written in Go. The theme, fonts and scripts are embedded in it, so there is nothing else to install on the server.

## Requirements

- Go 1.27 or newer, only to build it.
- Linux, macOS or Windows.

## Build

<Steps>
<Step title="Clone the repository">
```bash
git clone https://github.com/YutaViriato/Wavemark.git
cd Wavemark
```
</Step>
<Step title="Build the binary">
```bash
go build -o bin/wavemark ./cmd/wavemark
```
</Step>
<Step title="Serve a folder">
```bash
./bin/wavemark dev site
```

Open `http://localhost:4321`. Edit any file in `site/` and the page reloads by itself.
</Step>
</Steps>

## Development and production

<Tabs>
<Tab title="dev">
`wavemark dev` reloads the browser on every change and shows each page's errors in an overlay. Pages with errors stay visible so you can see the problem.

```bash
wavemark dev site --addr :4321
```
</Tab>
<Tab title="serve">
`wavemark serve` only starts when the content has no errors. Once it is up, a change with errors does not take the site down: each affected page keeps its last valid version.

```bash
wavemark serve site --addr :8080
```
</Tab>
</Tabs>

<Warning title="Behind a proxy">
`serve` does not terminate TLS. In production, put it behind a reverse proxy such as Caddy or nginx. See [Deployment](./deploy.md).
</Warning>

Next: [write your first page](./first-page.md).
