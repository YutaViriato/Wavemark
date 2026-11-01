---
title: Installation
description: Install the binary with one line and serve your first folder.
---

Wavemark is a single binary. The theme, fonts and scripts are embedded in it, so there is nothing else to install.

## Install

<Tabs>
<Tab title="Linux and macOS">
```bash
curl -fsSL https://yutaviriato.github.io/Wavemark/install.sh | sh
```

The binary goes to `~/.local/bin`. If that folder is not on your `PATH`, the script prints the line to add.
</Tab>
<Tab title="Windows">
```powershell
irm https://yutaviriato.github.io/Wavemark/install.ps1 | iex
```

The binary goes to `%LOCALAPPDATA%\wavemark`, which is added to your user `PATH`.
</Tab>
<Tab title="Manual">
Download the archive for your system from [GitHub Releases](https://github.com/YutaViriato/Wavemark/releases), check it against `checksums.txt` and put `wavemark` anywhere on your `PATH`.
</Tab>
</Tabs>

Both scripts verify the download's checksum. Set `WAVEMARK_VERSION=v0.1.0` to pin a version, or `WAVEMARK_INSTALL_DIR` to choose the folder. Running the script again upgrades Wavemark.

## Start a site

<Steps>
<Step title="Create it">
```bash
wavemark init my-site
```
</Step>
<Step title="Preview it">
```bash
wavemark dev my-site
```

Open `http://localhost:4321`. Edit any file in `my-site/` and the page reloads by itself.
</Step>
<Step title="Check it">
```bash
wavemark check my-site
```
</Step>
</Steps>

## Development and production

<Tabs>
<Tab title="dev">
`wavemark dev` reloads the browser on every change and shows each page's errors in an overlay. Pages with errors stay visible so you can see the problem.
</Tab>
<Tab title="build">
`wavemark build` writes the site as static HTML. It refuses to write anything when the content has errors.
</Tab>
<Tab title="serve">
`wavemark serve` runs the site as a live server. It only starts when the content has no errors, and a later change with errors does not take a page down: the page keeps its last valid version.
</Tab>
</Tabs>

See [Deployment](./deploy.md) for publishing.

## Build from source

You need Go 1.27 or newer.

```bash
git clone https://github.com/YutaViriato/Wavemark.git
cd Wavemark
go build -o bin/wavemark ./cmd/wavemark
```

Next: [write your first page](./first-page.md).
