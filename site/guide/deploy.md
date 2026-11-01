---
title: Deployment
description: Run wavemark serve behind a reverse proxy and keep it up.
---

In production, Wavemark is one long-running process. It reads the content folder, watches it, and serves HTTP. Put a reverse proxy in front of it for TLS and compression.

## The process

<Flow>
<FlowStep title="Proxy" icon="shield">TLS, compression</FlowStep>
<FlowStep title="wavemark serve" icon="server">:8080</FlowStep>
<FlowStep title="Content folder" icon="folder">git pull, rsync…</FlowStep>
</Flow>

`serve` refuses to start when the content has errors, so a broken deploy fails loudly instead of going live. After it starts, updating the folder is the deployment: pull new commits, and the server picks them up.

## A systemd unit

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

## Caddy in front

```text title="Caddyfile"
docs.example.com {
	encode zstd gzip
	reverse_proxy 127.0.0.1:8080
}
```

## What the server sends

<Columns cols={2}>
<Card title="Cache headers" icon="clock">
Theme assets have a content hash in their path and are cached for a year. Pages use `ETag` and revalidate.
</Card>
<Card title="Security headers" icon="lock">
A strict Content Security Policy, `nosniff` and a referrer policy on every response. Content SVGs run in a sandbox.
</Card>
</Columns>

<Tip title="Check in CI">
Run `wavemark check site --json` in your pipeline. It exits with status 1 when there are errors, so a pull request with a broken link fails before it reaches the server.
</Tip>
