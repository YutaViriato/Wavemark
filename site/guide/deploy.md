---
title: Deployment
description: Publish a site as static files on GitHub Pages, or run it as a live server.
---

There are two ways to publish a Wavemark site. Both refuse to publish content with errors.

<Columns cols={2}>
<Card title="Static files" icon="file">
`wavemark build` writes plain HTML. Host it on GitHub Pages, Netlify, Cloudflare Pages or any web server. No process to keep running.
</Card>
<Card title="Live server" icon="server">
`wavemark serve` runs the site as a process that watches the folder. Updating the folder is the deployment.
</Card>
</Columns>

## Static files

```bash
wavemark build my-site --out dist
```

The output folder has one `.html` file per page, the Markdown source of each page next to it, `404.html`, `llms.txt` and the theme assets. Pages are written as `guide/install.html`, so a host that serves `/guide/install` from `guide/install.html` works without configuration. GitHub Pages, Netlify and Cloudflare Pages all do.

If the site lives under a path, such as `https://you.github.io/my-repo/`, set `baseURL` in `wavemark.yaml` and every link gets the prefix:

```yaml title="wavemark.yaml"
baseURL: https://you.github.io/my-repo
```

`--base /my-repo` does the same from the command line.

### GitHub Pages

<Steps>
<Step title="Enable Pages">
In the repository settings, under Pages, set the source to **GitHub Actions**.
</Step>
<Step title="Add the workflow">
```yaml title=".github/workflows/pages.yml"
name: Pages
on:
  push:
    branches: [main]
permissions:
  contents: read
  pages: write
  id-token: write
jobs:
  deploy:
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - uses: actions/checkout@v4
      - run: curl -fsSL https://yutaviriato.github.io/Wavemark/install.sh | sh
      - run: ~/.local/bin/wavemark build my-site --out dist
      - uses: actions/upload-pages-artifact@v3
        with:
          path: dist
      - id: deployment
        uses: actions/deploy-pages@v4
```
</Step>
<Step title="Push">
Every push to `main` checks the content and publishes it. A push with errors fails the workflow and leaves the published site as it was.
</Step>
</Steps>

## Live server

<Flow>
<FlowStep title="Proxy" icon="shield">TLS, compression</FlowStep>
<FlowStep title="wavemark serve" icon="server">:8080</FlowStep>
<FlowStep title="Content folder" icon="folder">git pull, rsync…</FlowStep>
</Flow>

`serve` refuses to start when the content has errors, so a broken deploy fails loudly instead of going live. After it starts, pull new commits into the folder and the server picks them up. A page that starts having errors keeps its last valid version.

```ini title="/etc/systemd/system/wavemark.service"
[Unit]
Description=Wavemark
After=network.target

[Service]
ExecStart=/usr/local/bin/wavemark serve /srv/docs --addr 127.0.0.1:8080
Restart=on-failure
User=wavemark

[Install]
WantedBy=multi-user.target
```

```text title="Caddyfile"
docs.example.com {
	encode zstd gzip
	reverse_proxy 127.0.0.1:8080
}
```

Theme assets have a content hash in their path and are cached for a year. Pages use `ETag`. Every response has a strict Content Security Policy.

<Tip title="Check in CI">
Run `wavemark check my-site --json` in your pipeline. It exits with status 1 when there are errors, so a pull request with a broken link fails before it reaches anyone.
</Tip>
