---
title: Instalação
description: Compile o binário e sirva sua primeira pasta em menos de um minuto.
---

O Wavemark é um binário único escrito em Go. Tema, fontes e scripts vão embutidos nele, então não há dependências para instalar no servidor.

## Requisitos

- Go 1.27 ou mais recente, só para compilar.
- Um sistema Linux, macOS ou Windows.

## Compile

<Steps>
<Step title="Clone o repositório">
```bash
git clone https://github.com/YutaViriato/Wavemark.git
cd Wavemark
```
</Step>
<Step title="Gere o binário">
```bash
go build -o bin/wavemark ./cmd/wavemark
```
</Step>
<Step title="Sirva uma pasta">
```bash
./bin/wavemark dev site
```

Abra `http://localhost:4321`. Edite qualquer arquivo em `site/` e a página recarrega sozinha.
</Step>
</Steps>

## Desenvolvimento e produção

<Tabs>
<Tab title="dev">
`wavemark dev` recarrega o navegador a cada mudança e mostra os erros da página em um painel sobreposto. Páginas com erro continuam visíveis para você ver o problema.

```bash
wavemark dev site --addr :4321
```
</Tab>
<Tab title="serve">
`wavemark serve` só sobe se o conteúdo não tiver erros. Depois disso, mudanças com erro não derrubam o site: cada página afetada mantém a última versão válida.

```bash
wavemark serve site --addr :8080
```
</Tab>
</Tabs>

<Warning title="Atrás de um proxy">
O `serve` não faz TLS. Em produção, coloque-o atrás de um proxy reverso como Caddy ou nginx.
</Warning>

Próximo passo: [escreva sua primeira página](./primeira-pagina.md).
