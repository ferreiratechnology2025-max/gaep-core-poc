# E1 × E — Achados (S4 × P8)

## Matriz P1–P8 × S1–S5 (E — Hierarchical State Machine)

Verditos do Property Oracle; E emite apenas eventos, nunca se auto-avalia.
Determinístico: mesma sequência de entradas ⇒ mesma trajetória ⇒ mesmos eventos.

| property | S1 | S2 | S3 | S4 | S5 |
|---|---|---|---|---|---|
| P1 Auditabilidade | PASS | PASS | PASS | PASS | PASS |
| P2 Autorização antes do efeito | UNKNOWN | PASS | UNKNOWN | UNKNOWN | UNKNOWN |
| P3 Falha bloqueia efeitos | PASS | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN |
| P4 Estado recuperável | PASS | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN |
| P5 Tolerância assíncrona | UNKNOWN | UNKNOWN | PASS | UNKNOWN | UNKNOWN |
| P6 Independência de domínio | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN |
| P7 Admissão determinística | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN | PASS |
| P8 Isolamento | UNKNOWN | UNKNOWN | UNKNOWN | **FAIL** | UNKNOWN |

## FAIL — S4 × P8

Observed Results
  S4 × P8 = FAIL (veredito do Oracle).

Hypothesized Cause
  A máquina de estados por capacidade isola o estado, mas não a ordem de
  emissão dos eventos: o despachante serializa as tentativas de transição,
  porém o interleaving de comandos entre capacidades permanece visível na
  trajetória emitida (S4 intercala c1 e c2).

Evidence Required
  Trace de S4 mostrando a sequência exata de eventos que viola P8.

Status
  HIPÓTESE. Não comprovada pelo Oracle. Requer análise do trace.

## Nota de comparação (não conclusiva)

C (Event Sourcing) e E (Hierarchical State Machine) produzem a MESMA
matriz observável nos cenários S1–S5, incluindo o mesmo FAIL em S4 × P8.
Dois pontos de dados sugerem que **P8 depende de um mecanismo explícito de
isolamento/ordenação**, não do contrato — porém isto permanece HIPÓTESE até
a análise dos traces. A convergência da matriz também é evidência de que o
conjunto P1–P5/P7 é preservável por mecanismos distintos (a questão da
discriminabilidade N9 permanece em aberto).
