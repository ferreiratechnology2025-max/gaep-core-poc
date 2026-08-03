# Structural Cost Matrix (SCM)

> **O desempate entre arquiteturas equivalentes.** Quando duas ou mais
> arquiteturas preservam exatamente as mesmas propriedades (mesmos vereditos
> do Oracle em P1–P8) e honram os mesmos contratos observáveis, elas não são
> "a mesma coisa": podem diferir em custo estrutural. Sem esta matriz, o
> resultado permanece empatado indefinidamente e a decisão de qual vira
> Reference Implementation fica arbitrária. Esta matriz **não substitui** a de
> necessidade — **complementa** a decisão quando há múltiplos mecanismos
> equivalentes (fluxo: Property → arquiteturas que preservam → custo →
> trade-offs → decisão).

## 1. Dimensões de custo (v1)

| Dimensão | O que mede | Sinal do custo |
|---|---|---|
| **Complexidade estrutural** | nº de conceitos/componentes/estados distintos necessários para satisfazer o contrato | mais conceitos = maior |
| **Número de estados** | tamanho do espaço de estados observável (mínimo, não o representado) | mais estados = maior |
| **Acoplamento** | quantos componentes precisam conhecer o estado uns dos outros; quantidade de canais de comunicação | mais canais = maior |
| **Verificabilidade formal** | facilidade de modelar/verificar (modelagem direta em máquina de estados/automato vs. necessidade de modelar concorrência/mensagens) | mais difícil = maior |
| **Esforço de implementação** | linhas de código e código de suporte (locking, serialização, replay) para honrar o contrato | mais código = maior |
| **Superfície de manutenção** | nº de pontos onde uma mudança no contrato toca a arquitetura | mais pontos = maior |

**Medição honesta:** todas as dimensões são **qualitativas** (escala
Baixo/Médio/Alto) na v1. Nenhuma delas é critério de *correção* — só de
*desempate*. Sinalizar como estimativa, não como medição precisa.

## 2. Formato

```
Arquiteturas que preservam P1–P8: A, B′, C′

| Dimensão        | A (GAEP) | B′ (guardada) | C′ (event sourcing) |
|-----------------|----------|---------------|---------------------|
| Complexidade    | Alto     | Médio         | Médio               |
| Estados         | Médio    | Baixo         | Alto                |
| Acoplamento     | Médio    | Baixo         | Alto (replay)       |
| Verificabilidade| Médio    | Alto          | Médio               |
| Implementação   | Médio    | Baixo         | Alto                |
| Manutenção      | Médio    | Baixo         | Alto                |
```

Linha de **trade-off**: para cada par empatado, a linha descreve o que se
ganha e o que se paga.

## 3. Regras de desempate

1. **Só compara quem empatou nas propriedades.** A SCM entra apenas quando o
   Oracle devolve vereditos idênticos (P1–P8) para duas ou mais arquiteturas.
   Arquitetura que viola propriedade não entra na matriz de custo.
2. **Nenhuma dimensão é absoluta.** Não existe peso global fixo; a decisão é
   por trade-off explícito registrado na linha de decisão.
3. **Decisão registra o mecanismo escolhido como Reference Implementation,**
   **nunca promove o mecanismo a arquitetura.** O contrato observável
   (vocabulário + propriedades) permanece na Architecture Specification; o
   custo decide *qual implementação* o exemplifica.
4. **Reabertura:** se um novo mecanismo empata nas propriedades e vence em
   custo, a decisão anterior é reaberta (a SCM é hipótese como tudo mais).
5. **Custo não compensa propriedade.** Uma arquitetura com custo menor que
   viola P1–P8 nunca é escolhida por custo — esta matriz é só desempate.

## 4. Status e ligação

- **Status:** protocolo congelado; **matriz vazia** — nenhuma arquitetura
  concorrente completa (A–E) passou ainda pela bateria do Oracle. O preenchimento
  é a saída do E1 (`experiment-plan.md`).
- **Limitação explícita:** as seis dimensões são qualitativas na v1; se um
  desempate crítico depender delas, formalizar métricas numéricas como
  sub-experimento (contagem de conceitos, estados, pontos de acoplamento).
- **Relacionamento:** complementa `architectural-necessity-matrix.md`
  (fonte da verdade) e `observable-contract-specification.md` (o contrato que
  a SCM desempatou).
