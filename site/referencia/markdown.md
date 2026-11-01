---
title: Markdown
description: O Markdown aceito pelo Wavemark e como cada elemento aparece no tema.
---

O Wavemark aceita CommonMark com as extensões do GitHub (GFM): tabelas, texto riscado, listas de tarefas, autolinks e notas de rodapé.

## Texto

Parágrafos usam o corpo em sans de 15 px. **Negrito** fica com a cor de texto forte, *itálico* mantém a cor do corpo e ~~riscado~~ fica apagado. Código inline como `wavemark check` usa a fonte mono. Links internos como [a instalação](../guia/instalacao.md) e externos como [o CommonMark](https://commonmark.org) têm sublinhado discreto.[^1]

> Citações ficam recuadas, com uma borda à esquerda.

## Listas

- Itens de lista
- Com marcadores apagados
  - E níveis aninhados

1. Listas numeradas
2. Seguem o mesmo ritmo

- [x] Tarefa concluída
- [ ] Tarefa pendente

## Tabelas

| Comando | Para que serve | Sai com erro? |
|:---|:---|:---:|
| `wavemark dev` | Escrever com recarga automática | Não |
| `wavemark check` | Validar antes de publicar | Sim, se houver erros |
| `wavemark serve` | Produção | Não sobe com erros |

## Código

```go title="internal/site/site.go"
// Build scans the folder and produces a snapshot.
func Build(opts Options) (*Site, error) {
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, err
	}
	return &Site{Root: root}, nil
}
```

```json
{
  "ok": false,
  "errors": 1,
  "diagnostics": [{ "file": "guia/x.md", "line": 12, "code": "link.broken" }]
}
```

```
Bloco sem linguagem: texto puro, sem realce.
```

### Títulos de terceiro nível

Títulos `###` também entram no sumário, recuados.

#### Quarto nível

Títulos `####` usam sans em negrito e ficam fora do sumário.

---

[^1]: Notas de rodapé aparecem no fim da página.
