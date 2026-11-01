---
title: Configuração
description: Três níveis de configuração, do site inteiro até uma página, com herança.
---

A configuração tem três níveis. Cada nível mais específico sobrescreve o anterior:

1. `wavemark.yaml`, na raiz, vale para o site inteiro.
2. `_meta.yaml`, em cada pasta, vale para a pasta e suas subpastas.
3. O frontmatter vale só para a página.

## wavemark.yaml

```yaml title="wavemark.yaml"
title: Acme Docs
description: Documentação da Acme.
lang: pt-BR
nav:
  - label: Guia
    href: /guia
defaults:
  toc: true
```

As chaves `title`, `nav`, `logo` e `fonts` só existem aqui. A chave `defaults` define o que todas as páginas herdam.

## _meta.yaml

```yaml title="guia/_meta.yaml"
title: Guia
order: [index, instalacao, primeira-pagina]
defaults:
  layout: docs
```

`order` lista arquivos sem `.md` e nomes de subpastas. O que não estiver na lista vem depois, ordenado por `order` do frontmatter e depois por título.

## Cascata

<Accordion title="Exemplo de herança" open>
Com `toc: false` em `wavemark.yaml` e `toc: true` em `guia/_meta.yaml`, todas as páginas de `guia/` mostram o sumário e as demais não. Uma página dentro de `guia/` ainda pode escrever `toc: false` no frontmatter.
</Accordion>

<Accordion title="O que não herda">
`title`, `description` e `order` são da página. Só as chaves de `defaults` (`layout`, `toc` e `hidden`) passam de um nível para o outro.
</Accordion>

## Fontes próprias

O tema usa fontes livres por padrão. Se você tem licença para outras fontes, aponte os arquivos `.woff2` em `wavemark.yaml`:

```yaml
fonts:
  sans: fontes/minha-sans.woff2
  serif: fontes/minha-serif.woff2
```

<Danger title="Licença das fontes">
Não use as fontes da Anthropic sem licença. Elas são proprietárias.
</Danger>
