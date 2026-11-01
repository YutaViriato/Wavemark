---
title: Wavemark
description: Aponte para uma pasta de Markdown e ela vira um site.
layout: landing
---

<Hero eyebrow="Wavemark" title="Markdown que vira site enquanto você escreve">
Um binário, uma pasta de arquivos `.md` e nenhuma etapa de build. Feito para documentação longa escrita por pessoas e por agentes.

<Button href="./guia/instalacao.md" label="Começar" />
<Button href="./referencia/componentes.md" label="Ver componentes" variant="secondary" />
</Hero>

## Escreva Markdown. O resto é runtime.

<Columns cols={3}>
<Card title="Sem build" href="./guia/index.md" cta="Como funciona">
O servidor lê a pasta, guarda um snapshot em memória e reconstrói quando um arquivo muda.
</Card>
<Card title="Estrito de propósito" href="./agentes.md" cta="Para agentes">
Atributo desconhecido, link quebrado ou HTML cru viram erro com linha, coluna e dica de correção.
</Card>
<Card title="Componentes declarativos" href="./referencia/componentes.md" cta="Referência">
Cards, abas, passos e destaques em tags estilo JSX, sem JavaScript no conteúdo.
</Card>
</Columns>

## Do arquivo à página

<CodeGroup>
```bash title="Servidor de desenvolvimento"
wavemark dev site
```
```bash title="Validação"
wavemark check site --json
```
```bash title="Produção"
wavemark serve site --addr :8080
```
</CodeGroup>
