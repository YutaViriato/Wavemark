---
title: Visão geral
description: O que o Wavemark faz, como ele organiza uma pasta em site e o que acontece quando um arquivo muda.
---

Wavemark é um runtime: um processo que serve uma pasta de Markdown como site. Não existe etapa de build nem pasta `dist/`. Você edita um arquivo, o servidor percebe a mudança e a próxima requisição já recebe a página nova.

## Como uma pasta vira site

Cada arquivo `.md` é uma página e cada pasta é uma seção da navegação. A rota vem do caminho do arquivo:

| Arquivo | Rota |
|---|---|
| `index.md` | `/` |
| `guia/index.md` | `/guia` |
| `guia/instalacao.md` | `/guia/instalacao` |
| `referencia/componentes.md` | `/referencia/componentes` |

Imagens e outros arquivos são servidos no mesmo caminho em que estão na pasta.

## O que acontece quando um arquivo muda

O servidor mantém um *snapshot* imutável do site em memória: páginas renderizadas, navegação e índice de links. Quando o sistema de arquivos avisa uma mudança, o Wavemark espera 120 ms para agrupar salvamentos em sequência, reconstrói o snapshot e troca o antigo pelo novo de uma vez. Requisições em andamento nunca veem um site pela metade.

<Callout type="tip" title="Última versão válida">
Em produção, se uma página passar a ter erros, o servidor continua entregando a última versão válida dela e registra o problema no log. As outras páginas seguem atualizando normalmente.
</Callout>

## Por que tão estrito

O conteúdo do Wavemark costuma ser gerado por agentes de IA. Um agente não vê a página renderizada, então todo erro precisa aparecer como texto: arquivo, linha, coluna, código e uma dica de correção. Por isso:

- HTML cru é proibido. A identidade visual vem só dos componentes.
- Links internos são validados, inclusive as âncoras.
- Chaves de configuração desconhecidas são erro, não silêncio.

Veja [Para agentes](../agentes.md) para o fluxo completo de geração e validação.
