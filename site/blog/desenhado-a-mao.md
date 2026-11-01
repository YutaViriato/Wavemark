---
title: Desenhado à mão, gerado por código
description: Como os ícones do Wavemark ganham traço de caneta sem nenhum arquivo de imagem.
date: 2026-11-02
author: Equipe Wavemark
icon: pencil
---

Os ícones do Wavemark não são arquivos. Cada um é uma lista de formas simples (retas, polígonos, círculos, curvas) num quadrado de 48 por 48. Na hora de renderizar, o tema passa essas formas por uma caneta.

## A caneta

A caneta faz três coisas pequenas:

1. Desloca cada vértice alguns décimos de unidade, para que nenhum canto fique perfeito.
2. Curva levemente cada aresta, como uma mão que não é régua.
3. Passa um pouco do ponto inicial ao fechar uma forma, como quem fecha um círculo no papel.

O preenchimento usa a mesma forma, com outra perturbação e um pequeno deslocamento. O resultado lembra uma impressão um pouco fora de registro: a tinta e a cor não coincidem exatamente.

<Columns cols={4}>
<Card title="Caneta" icon="pencil">
Traço único.
</Card>
<Card title="Frasco" icon="flask">
Preenchimento.
</Card>
<Card title="Broto" icon="sprout">
Curvas.
</Card>
<Card title="Bússola" icon="compass">
Círculos.
</Card>
</Columns>

## Determinismo

A perturbação vem de um gerador pseudoaleatório com semente fixa, derivada do nome do ícone. O foguete é sempre o mesmo foguete. Isso mantém o `ETag` das páginas estável e permite comparar o HTML em testes.

## Movimento

Todo traço tem comprimento normalizado (`pathLength="1"`). Isso permite que o CSS anime o traço do zero ao fim com a mesma regra para qualquer ícone. Quando um cartão entra na tela, o ícone se desenha; os marcadores de `<Steps>` também.

<Quote author="Nota de projeto">
O movimento existe para mostrar que a página é feita de traços, não para chamar atenção. Quem pede menos movimento no sistema vê tudo pronto, sem animação.
</Quote>
