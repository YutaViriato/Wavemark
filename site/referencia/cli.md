---
title: CLI
description: Os comandos do binário wavemark.
---

## wavemark dev

```bash
wavemark dev [pasta] [--addr :4321]
```

Serve a pasta, recarrega o navegador a cada mudança e mostra os erros de cada página em um painel. O terminal lista todos os diagnósticos após cada reconstrução.

## wavemark serve

```bash
wavemark serve [pasta] [--addr :8080]
```

Servidor de produção. Recusa subir se houver erros. Depois de no ar, uma página que passa a ter erros continua com a última versão válida.

Assets do tema são servidos com hash no caminho e cache imutável. Páginas usam `ETag`.

## wavemark check

```bash
wavemark check [pasta] [--json]
```

Valida tudo e sai com código 1 se houver erros. Com `--json`, a saída é:

```json
{
  "ok": false,
  "pages": 9,
  "errors": 1,
  "warnings": 0,
  "diagnostics": [
    {
      "file": "guia/x.md",
      "line": 12,
      "column": 5,
      "severity": "error",
      "code": "component.attr-unknown",
      "message": "atributo desconhecido 'titl' em <Card>",
      "hint": "Você quis dizer 'title'? Atributos válidos: title, href, cta."
    }
  ]
}
```

Os códigos (`code`) são estáveis e podem ser usados por ferramentas.

## wavemark spec

```bash
wavemark spec [--json]
```

Imprime a referência de sintaxe, componentes e configuração. É gerada dos mesmos schemas que validam o conteúdo. Com `--json`, imprime os schemas dos componentes.
