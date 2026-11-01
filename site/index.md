---
title: Wavemark
description: Aponte para uma pasta de Markdown e ela vira um site.
layout: landing
---

<Hero eyebrow="Wavemark" title="Markdown que vira site enquanto você escreve" highlight="enquanto você escreve">
Um binário, uma pasta de arquivos `.md` e nenhuma etapa de build. Feito para documentação longa escrita por pessoas e por agentes.

<Button href="./guia/instalacao.md" label="Começar" />
<Button href="./referencia/componentes.md" label="Ver componentes" variant="secondary" />
</Hero>

## Escreva Markdown. O resto é runtime.

<Columns cols={3}>
<Card title="Sem build" icon="bolt" href="./guia/index.md" cta="Como funciona">
O servidor lê a pasta, guarda um snapshot em memória e reconstrói quando um arquivo muda.
</Card>
<Card title="Estrito de propósito" icon="shield" href="./agentes.md" cta="Para agentes">
Atributo desconhecido, link quebrado ou HTML cru viram erro com linha, coluna e dica de correção.
</Card>
<Card title="Componentes declarativos" icon="layers" href="./referencia/componentes.md" cta="Referência">
Cards, abas, passos e destaques em tags estilo JSX, sem JavaScript no conteúdo.
</Card>
</Columns>

<Columns cols={4}>
<Stat value="1" label="binário, com tema, fontes e scripts embutidos" />
<Stat value="0" label="etapas de build entre salvar e publicar" />
<Stat value="~0,5 s" label="para reconstruir 600 páginas" />
<Stat value="27" label="componentes, todos validados por schema" />
</Columns>

<Section tone="dark">
<Quote author="Equipe Wavemark" role="Princípio de projeto">
Se um agente não consegue ver a página, cada erro precisa existir como texto: arquivo, linha, coluna e o que fazer.
</Quote>
</Section>

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

## Do blog

<PageList dir="./blog" style="cards" limit={3} />

<Section tone="oat" align="center">
## Comece com uma pasta

Crie um arquivo `.md`, rode `wavemark dev` e abra o navegador.

<Button href="./guia/instalacao.md" label="Instalar o Wavemark" />
</Section>
