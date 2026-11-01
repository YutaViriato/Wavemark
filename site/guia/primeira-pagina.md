# Primeira página

Crie um arquivo `.md` em qualquer lugar da pasta. O primeiro título `#` vira o título da página e o resto é Markdown comum.

```markdown title="guia/ola.md"
# Olá

Esta é a minha página. Ela tem um [link para a instalação](./instalacao.md).

## Uma seção

Texto da seção.
```

A página fica disponível em `/guia/ola` e aparece na navegação lateral, dentro da seção Guia.

## Frontmatter

Para controlar título, descrição e posição na navegação, use um bloco YAML no topo:

```yaml title="Frontmatter"
---
title: Olá
description: Uma frase que aparece abaixo do título.
order: 2
---
```

Quando o frontmatter tem `title`, o corpo não pode ter `# H1`. A regra existe para que a página tenha um título só.

## Componentes

Componentes são tags com inicial maiúscula, cada uma em sua própria linha. O conteúdo entre as tags é Markdown:

```mdx title="Um destaque"
<Note title="Lembrete">
Rode `wavemark check` antes de publicar.
</Note>
```

<Note title="Lembrete">
Rode `wavemark check` antes de publicar.
</Note>

A lista completa está na [referência de componentes](../referencia/componentes.md).

## Links e âncoras

Aponte links para o arquivo, não para a URL. O Wavemark converte o caminho para a rota e avisa se o destino não existe:

- `[Instalação](./instalacao.md)` vira `/guia/instalacao`.
- `[Abas](../referencia/componentes.md#abas)` aponta para uma seção de outra página.
- Âncoras saem dos títulos sem acentos: `## Configuração avançada` vira `#configuracao-avancada`.
