# Questões de Pesquisa Abertas

## Princípio metodológico (permanente)

Toda nova fase experimental deve declarar explicitamente a pergunta de
pesquisa que pretende responder antes da elaboração de qualquer plano de
implementação. O roadmap é dirigido por "o que ainda não sabemos", não
por "o que ainda não testamos".

Este princípio protege o programa experimental contra expansão
especulativa: uma fase sem pergunta explícita não é autorizada.

## Estado atual do programa experimental

**Não há, neste momento, nenhuma pergunta de pesquisa ativa que justifique
a abertura de uma nova fase experimental.**

O programa está em estado de maturidade metodológica:
- não há backlog experimental;
- não há dívida metodológica registrada;
- não há experimento parcialmente concluído;
- não há evidência aguardando interpretação;
- não há decisão arquitetural pendente decorrente automaticamente dos resultados.

Hipóteses em aberto (H3 — sensibilidade de D ao escalonador) permanecem
como tal por escolha metodológica: sua investigação requer replicação
cross-platform ou análise causal de trace, o que não é prioridade até que
uma pergunta de pesquisa específica justifique o esforço.

## Perguntas encerradas

| Pergunta | Experimento | Resultado |
|---|---|---|
| A bateria S1–S5 discrimina mecanismos? | E1 | Parcialmente — 1/40 células (P8 via S4) |
| Cobertura expandida sem regressão é possível? | E2.1 | Sim — S6–S10 adicionados, sem regressão |
| Aumentar cobertura aumenta discriminação? | E2.2 | Sim — duas famílias independentes (P8 isolamento, P3 rollback) |
| Divergências são achados ou defeitos? | E2.3 | Achados — classificados, não corrigidos |

## Perguntas rejeitadas

| Pergunta | Razão da rejeição |
|---|---|
| Executar A para completar a matriz | Completude documental, não ganho epistemológico — A permanece não comparada por decisão metodológica |
