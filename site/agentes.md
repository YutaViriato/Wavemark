---
title: Para agentes
description: Como um agente de IA deve gerar e validar conteúdo para o Wavemark.
order: 10
---

O Wavemark foi desenhado para conteúdo gerado por agentes. O agente não vê a página renderizada, então o runtime transforma cada problema em um diagnóstico de texto.

## O ciclo

<Steps>
<Step title="Leia a referência">
Rode `wavemark spec` e coloque a saída no contexto. Ela descreve todos os componentes, atributos e regras, e é gerada do próprio código.
</Step>
<Step title="Escreva os arquivos">
Crie os `.md` na pasta do site. Aponte links para arquivos (`./outra.md`), nunca para URLs inventadas.
</Step>
<Step title="Valide">
Rode `wavemark check site --json`. Cada diagnóstico tem arquivo, linha, coluna, código e, quase sempre, uma dica.
</Step>
<Step title="Corrija até zerar">
Repita até `"ok": true`. Avisos não bloqueiam, mas vale resolvê-los.
</Step>
</Steps>

## Erros comuns

| Código | Causa | Correção |
|---|---|---|
| `html.forbidden` | `<br>`, `<div>` ou outro HTML | Use Markdown ou um componente |
| `component.unknown` | `<CardGroup>` e outros nomes do Mintlify | Siga a dica: `<Columns>` |
| `component.attr-value` | `cols="2"` | `cols={2}` |
| `link.broken` | Link para arquivo inexistente | Use o caminho do `.md` relativo à página |
| `link.anchor` | `#secao` que não existe | Âncoras vêm dos títulos, sem acentos |
| `title.duplicate` | `title:` no frontmatter e `# H1` no corpo | Escolha um |

<Info title="Por que não MDX">
MDX executa JavaScript e aceita qualquer HTML. Isso torna o conteúdo impossível de validar e fácil de quebrar visualmente. O Wavemark usa a mesma aparência de tags, mas cada tag tem um schema.
</Info>
