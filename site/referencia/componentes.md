---
title: Componentes
description: Todos os componentes, com o código e o resultado lado a lado.
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

<Success title="Pronto">
`<Success>` (ou `<Check>`) confirma que algo deu certo.
</Success>

<Danger title="Perigo">
A ação não pode ser desfeita.
</Danger>

## Cards e colunas

`<Columns>` cria uma grade de 1 a 4 colunas (`<CardGroup>` é um atalho com o mesmo efeito). Em telas estreitas ela vira uma coluna só. Um `<Card>` com `href` vira um link inteiro, e `icon` acrescenta um [ícone desenhado à mão](#icones).

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
<Card title="Instalação" icon="rocket" href="../guia/instalacao.md" cta="Instalar">
Compile o binário e sirva uma pasta.
</Card>
<Card title="Configuração" icon="gear" href="../guia/configuracao.md" cta="Configurar">
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

Para uma lista de perguntas, `<AccordionGroup>` junta os itens em um bloco só:

<AccordionGroup>
<Accordion title="O Wavemark gera site estático?">
Ainda não. Ele é um runtime: serve a pasta diretamente.
</Accordion>
<Accordion title="Funciona sem JavaScript no navegador?">
Sim. Abas mostram todo o conteúdo, e o resto funciona normalmente.
</Accordion>
<Accordion title="Respeita quem prefere menos movimento?">
Sim. Com a preferência do sistema ligada, nenhuma animação roda.
</Accordion>
</AccordionGroup>

## Botões

Botões em sequência ficam lado a lado. O texto vem de `label` ou do conteúdo.

```mdx
<Button href="../guia/instalacao.md" label="Começar" />
<Button href="https://github.com/YutaViriato/Wavemark" variant="secondary">Código-fonte</Button>
```

<Button href="../guia/instalacao.md" label="Começar" />
<Button href="https://github.com/YutaViriato/Wavemark" variant="secondary">Código-fonte</Button>

## Citações

<Quote author="Ada Lovelace" role="Notas sobre a Máquina Analítica, 1843">
A Máquina Analítica não tem pretensão alguma de originar nada. Ela pode fazer qualquer coisa que saibamos ordenar que faça.
</Quote>

## Números

<Columns cols={3}>
<Stat value="12" label="páginas neste site" />
<Stat value="52" label="ícones desenhados à mão" />
<Stat value="0" label="linhas de HTML cru" />
</Columns>

## Etiquetas e ícones no texto

`<Badge>` e `<Icon />` são inline: ficam no meio de parágrafos, títulos e tabelas, e abrem e fecham na mesma linha.

```mdx
Exportação estática <Badge color="clay">Em breve</Badge> e busca <Badge>Planejado</Badge>.
Feito com <Icon name="heart" label="carinho" /> e <Icon name="coffee" label="café" />.
```

Exportação estática <Badge color="clay">Em breve</Badge> e busca <Badge>Planejado</Badge>. Cores: <Badge>gray</Badge> <Badge color="clay">clay</Badge> <Badge color="olive">olive</Badge> <Badge color="sky">sky</Badge> <Badge color="fig">fig</Badge>.

Feito com <Icon name="heart" label="carinho" /> e <Icon name="coffee" label="café" />.

## Listas de páginas

`<PageList />` lista as páginas de uma pasta, com data, autor e descrição. Com `sort="auto"` (o padrão), páginas com data vêm das mais recentes para as mais antigas.

```mdx
<PageList dir="../blog" />
```

<PageList dir="../blog" />

## Ícones

Use o nome em `<Card icon>`, `<Icon name>` ou na chave `icon` do frontmatter. Nomes comuns do Lucide e do Font Awesome também funcionam, como `rocket-launch`, `settings` e `magnifying-glass`.

<Columns cols={4}>
<Card title="bolt" icon="bolt">
`icon="bolt"`
</Card>
<Card title="book" icon="book">
`icon="book"`
</Card>
<Card title="box" icon="box">
`icon="box"`
</Card>
<Card title="brush" icon="brush">
`icon="brush"`
</Card>
<Card title="calendar" icon="calendar">
`icon="calendar"`
</Card>
<Card title="chart" icon="chart">
`icon="chart"`
</Card>
<Card title="chat" icon="chat">
`icon="chat"`
</Card>
<Card title="check" icon="check">
`icon="check"`
</Card>
<Card title="clock" icon="clock">
`icon="clock"`
</Card>
<Card title="cloud" icon="cloud">
`icon="cloud"`
</Card>
<Card title="code" icon="code">
`icon="code"`
</Card>
<Card title="coffee" icon="coffee">
`icon="coffee"`
</Card>
<Card title="compass" icon="compass">
`icon="compass"`
</Card>
<Card title="cpu" icon="cpu">
`icon="cpu"`
</Card>
<Card title="database" icon="database">
`icon="database"`
</Card>
<Card title="download" icon="download">
`icon="download"`
</Card>
<Card title="eye" icon="eye">
`icon="eye"`
</Card>
<Card title="file" icon="file">
`icon="file"`
</Card>
<Card title="flag" icon="flag">
`icon="flag"`
</Card>
<Card title="flask" icon="flask">
`icon="flask"`
</Card>
<Card title="folder" icon="folder">
`icon="folder"`
</Card>
<Card title="gear" icon="gear">
`icon="gear"`
</Card>
<Card title="globe" icon="globe">
`icon="globe"`
</Card>
<Card title="heart" icon="heart">
`icon="heart"`
</Card>
<Card title="home" icon="home">
`icon="home"`
</Card>
<Card title="key" icon="key">
`icon="key"`
</Card>
<Card title="layers" icon="layers">
`icon="layers"`
</Card>
<Card title="lightbulb" icon="lightbulb">
`icon="lightbulb"`
</Card>
<Card title="lock" icon="lock">
`icon="lock"`
</Card>
<Card title="mail" icon="mail">
`icon="mail"`
</Card>
<Card title="map" icon="map">
`icon="map"`
</Card>
<Card title="message" icon="message">
`icon="message"`
</Card>
<Card title="pencil" icon="pencil">
`icon="pencil"`
</Card>
<Card title="pin" icon="pin">
`icon="pin"`
</Card>
<Card title="play" icon="play">
`icon="play"`
</Card>
<Card title="plug" icon="plug">
`icon="plug"`
</Card>
<Card title="puzzle" icon="puzzle">
`icon="puzzle"`
</Card>
<Card title="question" icon="question">
`icon="question"`
</Card>
<Card title="rocket" icon="rocket">
`icon="rocket"`
</Card>
<Card title="search" icon="search">
`icon="search"`
</Card>
<Card title="server" icon="server">
`icon="server"`
</Card>
<Card title="shield" icon="shield">
`icon="shield"`
</Card>
<Card title="sparkles" icon="sparkles">
`icon="sparkles"`
</Card>
<Card title="sprout" icon="sprout">
`icon="sprout"`
</Card>
<Card title="star" icon="star">
`icon="star"`
</Card>
<Card title="tag" icon="tag">
`icon="tag"`
</Card>
<Card title="target" icon="target">
`icon="target"`
</Card>
<Card title="terminal" icon="terminal">
`icon="terminal"`
</Card>
<Card title="user" icon="user">
`icon="user"`
</Card>
<Card title="users" icon="users">
`icon="users"`
</Card>
<Card title="wand" icon="wand">
`icon="wand"`
</Card>
<Card title="wave" icon="wave">
`icon="wave"`
</Card>
</Columns>

## Figuras

<Figure src="../img/arquitetura.svg" alt="Diagrama: a pasta de Markdown passa pelo parser, vira um snapshot em memória e é servida por HTTP" caption="Do arquivo à resposta HTTP" />

## Só na landing

`<Hero>` abre a página, com título grande que entra palavra por palavra. `highlight` sublinha um trecho com um traço desenhado à mão; `image` coloca uma imagem ao lado. `<Section tone="muted|oat|clay|dark">` cria faixas de largura total. Veja a [página inicial](../index.md).

```mdx
<Hero title="Markdown que vira site" highlight="vira site" eyebrow="Wavemark">
Subtítulo em Markdown.

<Button href="/guia" label="Começar" />
</Hero>

<Section tone="dark">
<Quote author="Alguém">Uma frase.</Quote>
</Section>
```
