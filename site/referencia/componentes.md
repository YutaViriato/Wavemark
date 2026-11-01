---
title: Componentes
description: Todos os componentes da v1, com o código e o resultado lado a lado.
---

Componentes são tags com nome em PascalCase. Cada tag fica em uma linha própria e o conteúdo entre elas é Markdown. A referência gerada a partir dos schemas está em `wavemark spec`.

## Destaques

Use `<Callout>` com `type`, ou um dos atalhos: `<Note>`, `<Tip>`, `<Info>`, `<Warning>` e `<Danger>`.

```mdx
<Tip title="Atalho">
`<Tip>` é o mesmo que `<Callout type="tip">`.
</Tip>
```

<Note>
Uma nota neutra, sem título. Serve para observações que não mudam o resultado.
</Note>

<Tip title="Atalho">
`<Tip>` é o mesmo que `<Callout type="tip">`.
</Tip>

<Info title="Informação">
Use para contexto que ajuda, mas que o leitor pode pular.
</Info>

<Warning title="Atenção">
Algo pode dar errado se o leitor ignorar este aviso.
</Warning>

<Danger title="Perigo">
A ação não pode ser desfeita.
</Danger>

## Cards e colunas

`<Columns>` cria uma grade de 1 a 4 colunas. Em telas estreitas ela vira uma coluna só. Um `<Card>` com `href` vira um link inteiro.

```mdx
<Columns cols={2}>
<Card title="Instalação" href="../guia/instalacao.md" cta="Instalar">
Compile o binário e sirva uma pasta.
</Card>
<Card title="Configuração" href="../guia/configuracao.md" cta="Configurar">
Três níveis de configuração com herança.
</Card>
</Columns>
```

<Columns cols={2}>
<Card title="Instalação" href="../guia/instalacao.md" cta="Instalar">
Compile o binário e sirva uma pasta.
</Card>
<Card title="Configuração" href="../guia/configuracao.md" cta="Configurar">
Três níveis de configuração com herança.
</Card>
</Columns>

<Columns cols={3}>
<Card title="Sem link">
Um cartão sem `href` é só um bloco de conteúdo.
</Card>
<Card title="Markdown dentro">
- Listas
- **Ênfase**
- `código`
</Card>
<Card title="Altura igual">
Cartões da mesma linha ficam com a mesma altura.
</Card>
</Columns>

## Abas

```mdx
<Tabs>
<Tab title="macOS">
Instale com Homebrew.
</Tab>
<Tab title="Linux">
Baixe o binário.
</Tab>
</Tabs>
```

<Tabs>
<Tab title="macOS">
Instale com Homebrew e confira a versão:

```bash
brew install wavemark
wavemark help
```
</Tab>
<Tab title="Linux">
Baixe o binário e coloque-o no `PATH`.
</Tab>
<Tab title="Windows">
Use o executável `wavemark.exe`.
</Tab>
</Tabs>

## Grupos de código

`<CodeGroup>` aceita só blocos de código cercados. O rótulo de cada aba vem de `title="..."` na cerca ou, sem ele, do nome da linguagem.

<CodeGroup>
```python title="main.py"
import anthropic

client = anthropic.Anthropic()
message = client.messages.create(
    model="claude-opus-5-5",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Olá, Claude"}],
)
print(message.content)
```
```typescript title="main.ts"
import Anthropic from "@anthropic-ai/sdk";

const client = new Anthropic();
const message = await client.messages.create({
  model: "claude-opus-5-5",
  max_tokens: 1024,
  messages: [{ role: "user", content: "Olá, Claude" }],
});
console.log(message.content);
```
```go title="main.go"
package main

import "fmt"

func main() {
	fmt.Println("Olá, Claude")
}
```
</CodeGroup>

## Passos

<Steps>
<Step title="Escreva">
Crie um arquivo `.md` na pasta do site.
</Step>
<Step title="Valide">
Rode `wavemark check` e corrija os erros.
</Step>
<Step title="Publique">Rode `wavemark serve`.</Step>
</Steps>

## Seções recolhíveis

<Accordion title="Posso usar HTML?">
Não. HTML cru é recusado para que o tema continue consistente. Se faltar um componente, ele deve entrar no Wavemark.
</Accordion>

<Accordion title="E JavaScript nos componentes?">
Também não. Valores de atributo são literais: textos, números, booleanos ou JSON.
</Accordion>

## Botões

Botões em sequência ficam lado a lado.

<Button href="../guia/instalacao.md" label="Começar" />
<Button href="https://github.com/YutaViriato/Wavemark" label="Código-fonte" variant="secondary" />

## Figuras

<Figure src="../img/arquitetura.svg" alt="Diagrama: a pasta de Markdown passa pelo parser, vira um snapshot em memória e é servida por HTTP" caption="Do arquivo à resposta HTTP" />

## Hero

`<Hero>` só existe no layout `landing`. Veja a [página inicial](../index.md) para um exemplo.
