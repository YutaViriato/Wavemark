---
title: CLI
description: The commands of the wavemark binary.
---

## wavemark dev

```bash
wavemark dev [folder] [--addr :4321]
```

Serves the folder, reloads the browser on every change and shows each page's errors in an overlay. The terminal lists every diagnostic after each rebuild. When the server restarts with a new binary, open pages reload by themselves.

## wavemark serve

```bash
wavemark serve [folder] [--addr :8080]
```

The production server. It refuses to start when there are errors. Once it is up, a page that starts having errors keeps its last valid version.

Theme assets are served with a hash in their path and an immutable cache. Pages use `ETag`. A `404.md` at the root, if present, is the not-found page.

## wavemark check

```bash
wavemark check [folder] [--json]
```

Validates everything and exits with status 1 when there are errors. With `--json`, the output is:

```json
{
  "ok": false,
  "pages": 9,
  "errors": 1,
  "warnings": 0,
  "diagnostics": [
    {
      "file": "guide/x.md",
      "line": 12,
      "column": 5,
      "severity": "error",
      "code": "component.attr-unknown",
      "message": "unknown attribute 'titl' on <Card>",
      "hint": "Did you mean 'title'? Valid attributes: title, href, icon, cta."
    }
  ]
}
```

The `code` values are stable and safe for tools to depend on.

## wavemark spec

```bash
wavemark spec [--json]
```

Prints the reference of the syntax, components and configuration. It is generated from the same schemas that validate content. With `--json`, it prints the component schemas, the icon names and the layouts.
