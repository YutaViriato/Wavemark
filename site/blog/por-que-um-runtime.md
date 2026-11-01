---
title: Por que um runtime
description: Servir a pasta diretamente, sem build, muda o que significa publicar.
date: 2026-11-01
author: Equipe Wavemark
icon: bolt
---

Geradores de site estático resolvem um problema real: transformar texto em HTML uma vez, e servir o resultado de qualquer lugar. Mas eles trazem uma etapa que ninguém pediu. Você salva, roda o build, espera, confere a pasta de saída e só então vê o que escreveu.

O Wavemark tira essa etapa do caminho. O processo que serve o site é o mesmo que lê os arquivos. Quando um `.md` muda, o servidor reconstrói o snapshot em memória e a próxima requisição já recebe a página nova.

## O snapshot

Cada reconstrução produz um snapshot imutável: páginas renderizadas, navegação, índice de rotas e de âncoras. A troca acontece de uma vez, com um ponteiro atômico. Uma requisição nunca vê metade de um site.

> Publicar deixa de ser um evento e vira um estado: o que está na pasta é o que está no ar.

Isso só é seguro porque o runtime é estrito. Um link quebrado não chega à produção, porque o servidor se recusa a subir com erros. Depois de no ar, uma página que passa a ter um erro continua servindo a última versão válida.

## O custo

Um runtime precisa de um processo rodando, e isso é um custo que um site estático não tem. Para documentação interna, bases de conhecimento e sites que mudam o tempo todo, a troca compensa. Para um site que muda uma vez por ano, talvez não.

<Figure src="../img/arquitetura.svg" alt="Diagrama: a pasta de Markdown passa pelo parser, vira um snapshot em memória e é servida por HTTP" caption="Do arquivo à resposta HTTP." />

## O que vem depois

Uma exportação estática é barata de acrescentar, porque a renderização já é uma função pura do snapshot. Ela vai entrar quando alguém precisar dela.
